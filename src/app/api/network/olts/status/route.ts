import { NextResponse } from 'next/server';
import { prisma } from '@/server/db/client';

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { oltIds } = body;

    if (!oltIds || !Array.isArray(oltIds) || oltIds.length === 0) {
      return NextResponse.json({ statusMap: {} });
    }

    // For each OLT, check if we have recent ONU data in the database
    // This gives us a rough "online/offline" status based on whether we can scan
    const statusMap: Record<string, { is_online: boolean; onu_count: number; online_count: number; last_scan?: string }> = {};

    for (const oltId of oltIds) {
      // Query the olt_onus table for cached data
      const onuCount = await prisma.$queryRawUnsafe<any[]>(
        `SELECT COUNT(*) as total, 
                SUM(CASE WHEN status = 'Online' THEN 1 ELSE 0 END) as online_count,
                MAX(scanned_at) as last_scan
         FROM olt_onus WHERE olt_id = ?`, oltId
      );

      if (onuCount && onuCount.length > 0) {
        const row = onuCount[0];
        const total = Number(row.total) || 0;
        const onlineCount = Number(row.online_count) || 0;
        const lastScan = row.last_scan;

        statusMap[oltId] = {
          is_online: total > 0,
          onu_count: total,
          online_count: onlineCount,
          last_scan: lastScan ? new Date(lastScan).toISOString() : undefined,
        };
      } else {
        statusMap[oltId] = {
          is_online: false,
          onu_count: 0,
          online_count: 0,
        };
      }
    }

    return NextResponse.json({ statusMap });
  } catch (error: any) {
    console.error('OLT status check error:', error);
    return NextResponse.json({ statusMap: {} });
  }
}
