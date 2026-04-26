import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { Prisma } from "@prisma/client";
import { nanoid } from "nanoid";
import { nowWIB } from "@/lib/timezone";

/**
 * RADIUS Post-Auth Hook
 * Called after successful authentication to:
 * 1. Set firstLoginAt and expiresAt on first login
 * 2. Check if voucher is expired
 * 3. Update voucher status
 * 
 * STRATEGY: Use nowWIB() to store times in WIB-as-UTC pattern
 * All datetime fields store WIB value in a UTC column (Prisma reads raw bytes)
 * nowWIB() = new Date(Date.now() + 7h offset) gives WIB wall-clock time as a Date object
 */

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { username, reply } = body;

    // Only process Access-Accept
    if (reply !== "Access-Accept") {
      return NextResponse.json({});
    }

    // Find voucher
    const voucher = await prisma.hotspotVoucher.findUnique({
      where: { code: username },
      include: { profile: true },
    });

    // If voucher not found in hotspotVoucher table, it might be:
    // - A PPPoE user (should not reach here due to conditional in FreeRADIUS)
    // - A legacy/test voucher in radcheck only
    // Return success to allow authentication to proceed
    if (!voucher) {
      return NextResponse.json({});
    }

    // Get current WIB time stored as WIB-as-UTC (matches the rest of the app's timezone pattern)
    const now = nowWIB();
    
    // Check if voucher is already expired (compare in same timezone)
    if (voucher.expiresAt && now > voucher.expiresAt) {
      // Mark as expired and reject
      await prisma.hotspotVoucher.update({
        where: { id: voucher.id },
        data: { status: "EXPIRED" },
      });

      return NextResponse.json(
        {
          success: false,
          error: "Voucher expired",
          action: "reject",
        },
        { status: 403 },
      );
    }

    // First login: set firstLoginAt and calculate expiresAt
    if (!voucher.firstLoginAt) {
      const { validityValue, validityUnit } = voucher.profile;

      // Calculate interval in milliseconds
      let intervalMs = 0;
      switch (validityUnit) {
        case "MINUTES":
          intervalMs = validityValue * 60 * 1000;
          break;
        case "HOURS":
          intervalMs = validityValue * 60 * 60 * 1000;
          break;
        case "DAYS":
          intervalMs = validityValue * 24 * 60 * 60 * 1000;
          break;
        case "MONTHS":
          // Approximate 30 days per month
          intervalMs = validityValue * 30 * 24 * 60 * 60 * 1000;
          break;
      }

      // Calculate expiresAt (server local time)
      const expiresAt = new Date(now.getTime() + intervalMs);

      // Update using Prisma - store as server local time (WIB)
      const updated = await prisma.hotspotVoucher.update({
        where: { id: voucher.id },
        data: {
          firstLoginAt: now,
          expiresAt: expiresAt,
          status: "ACTIVE",
        },
        select: { firstLoginAt: true, expiresAt: true },
      });

      // Auto-sync to Keuangan (realtime for manual/agent vouchers)
      if (!voucher.orderId) {
        try {
          // Use ID lookup (stable) - category name is "Penjualan Voucher Hotspot"
          const hotspotCategory = await prisma.transactionCategory.findUnique({
            where: { id: "cat-income-hotspot" },
          });

          if (hotspotCategory) {
            const existingTransaction = await prisma.transaction.findFirst({
              where: { reference: `VOUCHER-${voucher.code}` },
            });

            if (!existingTransaction) {
              // Check if this is an agent voucher
              const isAgentVoucher = voucher.agentId !== null;
              const hasResellerFee = voucher.profile.resellerFee > 0;
              
              // Income = sellingPrice (harga jual ke customer)
              const incomeAmount = voucher.profile.sellingPrice;
              
              // Create income transaction with selling price
              await prisma.transaction.create({
                data: {
                  id: nanoid(),
                  categoryId: hotspotCategory.id,
                  type: "INCOME",
                  amount: incomeAmount,
                  description: `Voucher ${voucher.profile.name} - ${voucher.code}${isAgentVoucher ? ' (Agent)' : ''}`,
                  date: now,
                  reference: `VOUCHER-${voucher.code}`,
                  notes: `Pendapatan voucher hotspot (Harga Jual: Rp ${incomeAmount}, Harga Modal: Rp ${voucher.profile.costPrice})`,
                },
              });
              console.log(
                `[POST-AUTH] Keuangan synced: ${voucher.code} - Income Rp ${incomeAmount}`,
              );

              // If agent voucher, record commission as expense
              // Net profit = sellingPrice - resellerFee
              if (isAgentVoucher && hasResellerFee) {
                const agentCategory = await prisma.transactionCategory.findUnique({
                  where: { id: "cat-expense-komisi" },
                });

                if (agentCategory) {
                  // Get agent name if available
                  const agent = await prisma.agent.findUnique({
                    where: { id: voucher.agentId! },
                    select: { name: true },
                  });
                  
                  const agentName = agent?.name || 'Unknown';
                  const commissionAmount = voucher.profile.resellerFee;
                  const netProfit = incomeAmount - commissionAmount;
                  
                  await prisma.transaction.create({
                    data: {
                      id: nanoid(),
                      categoryId: agentCategory.id,
                      type: "EXPENSE",
                      amount: commissionAmount,
                      description: `Komisi Agent ${agentName} - Voucher ${voucher.code}`,
                      date: now,
                      reference: `COMMISSION-${voucher.code}`,
                      notes: `Komisi agent untuk voucher ${voucher.profile.name} (Net Profit: Rp ${netProfit})`,
                    },
                  });
                  console.log(
                    `[POST-AUTH] Agent commission: ${voucher.code} - Rp ${commissionAmount} (Net: Rp ${netProfit})`,
                  );
                }
              }
            }
          }
        } catch (keuanganError) {
          console.error("[POST-AUTH] Keuangan sync error:", keuanganError);
        }
      }

      return NextResponse.json({});
    }

    // Subsequent logins: just verify not expired
    return NextResponse.json({});
  } catch (error: any) {
    console.error("RADIUS post-auth error:", error);
    return NextResponse.json({}, { status: 500 });
  }
}
