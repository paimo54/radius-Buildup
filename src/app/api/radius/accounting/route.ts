import { NextRequest, NextResponse } from "next/server";

/**
 * RADIUS Accounting Hook
 * Called by FreeRADIUS REST module for every accounting packet:
 *  - Acct-Status-Type = Start   → user mulai sesi (connect)
 *  - Acct-Status-Type = Stop    → user selesai sesi (disconnect)
 *  - Acct-Status-Type = Interim-Update → update statistik sesi
 *
 * All session data is stored in radacct by FreeRADIUS SQL module.
 * This endpoint just logs for debugging.
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const {
      username,
      statusType,
      nasIp,
      sessionTime,
      inputOctets,
      outputOctets,
    } = body;

    if (!username || !statusType) {
      return NextResponse.json({ success: true, action: "ignore" });
    }

    const normalizedStatus = statusType?.toLowerCase();

    if (normalizedStatus === "start") {
      console.log(`[ACCOUNTING] START: ${username} from ${nasIp}`);
    } else if (normalizedStatus === "stop") {
      console.log(
        `[ACCOUNTING] STOP: ${username} | session ${sessionTime}s | ` +
        `in: ${inputOctets}B out: ${outputOctets}B`
      );
    } else if (normalizedStatus === "interim-update" || normalizedStatus === "alive") {
      // FreeRADIUS SQL module handles radacct UPDATE directly — nothing else needed
    }

    return NextResponse.json({ success: true, action: normalizedStatus });
  } catch (error: any) {
    console.error("[ACCOUNTING] Error:", error);
    return NextResponse.json({ success: true, action: "error_ignored" });
  }
}
