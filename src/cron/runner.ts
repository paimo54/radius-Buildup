/**
 * Standalone Cron Runner — src/cron/runner.ts
 *
 * Menjalankan semua cron jobs langsung dari src/server/jobs/ tanpa HTTP.
 * Dijalankan via: tsx src/cron/runner.ts
 * PM2: pm2 start "tsx src/cron/runner.ts" --name salfanet-cron
 *
 * Keuntungan vs cron-service.js (HTTP polling):
 *  - Tidak bergantung pada Next.js server berjalan
 *  - Tidak ada network overhead / retry loop
 *  - Jobs tetap jalan saat Next.js restart/update
 *  - Error lebih cepat terdeteksi (stack trace lengkap)
 */

import 'dotenv/config';
import cron from 'node-cron';
import { CRON_JOBS } from '@/server/jobs/jobs.config';

// Track running jobs to prevent overlap
const runningJobs = new Set<string>();

const LOCK_JOBS = new Set([
  'invoice_generate',
  'auto_renewal',
  'pppoe_auto_isolir',
  'pppoe_session_sync',
  'telegram_backup',
  'webhook_log_cleanup',
  'activity_log_cleanup',
  'cron_history_cleanup',
]);

async function runJob(type: string, description: string): Promise<void> {
  if (runningJobs.has(type)) {
    console.log(`[CRON] Skipping ${description} — already running`);
    return;
  }

  const job = CRON_JOBS.find(j => j.type === type);
  if (!job) {
    console.warn(`[CRON] Job type not found: ${type}`);
    return;
  }

  if (LOCK_JOBS.has(type)) runningJobs.add(type);
  const start = Date.now();

  try {
    console.log(`[CRON] Starting: ${description}`);
    const result = await job.handler();
    const elapsed = ((Date.now() - start) / 1000).toFixed(1);
    console.log(`[CRON] Done: ${description} (${elapsed}s)`, result?.success !== false ? '✓' : '✗');
  } catch (err: any) {
    const elapsed = ((Date.now() - start) / 1000).toFixed(1);
    console.error(`[CRON] Error: ${description} (${elapsed}s)`, err?.message ?? err);
  } finally {
    runningJobs.delete(type);
  }
}

// ==================== STARTUP SEQUENCE ====================

// 1. FreeRADIUS health check — seed isolir radgroupreply sebelum job lain
setTimeout(() => runJob('freeradius_health', 'FreeRADIUS Health Check (startup)'), 5_000);

// 2. PPPoE auto isolir — catch expired users yang missed saat restart
setTimeout(() => runJob('pppoe_auto_isolir', 'PPPoE Auto Isolir (startup)'), 15_000);

// 3. Session recovery — reopen sessions yang salah ditutup saat update
setTimeout(async () => {
  try {
    // Langsung akses DB untuk session recovery tanpa bergantung pada job config
    const { prisma } = await import('@/server/db/client');
    const sixtyMinutesAgo = new Date(Date.now() - 60 * 60 * 1000);
    const result = await prisma.$executeRaw`
      UPDATE radacct
      SET acctstoptime = NULL, acctterminatecause = ''
      WHERE acctterminatecause = 'Lost-Carrier'
        AND acctstoptime >= ${sixtyMinutesAgo}
    `;
    console.log(`[CRON] Session recovery (startup): ${result} sessions restored`);
  } catch (err: any) {
    console.error('[CRON] Session recovery (startup) failed:', err?.message);
  }
}, 30_000);

// ==================== CRON SCHEDULES ====================

// Hanya schedule jobs yang punya fixed cron pattern (bukan 'dynamic')
const schedulableJobs = CRON_JOBS.filter(j => j.enabled && j.schedule !== 'dynamic');

for (const job of schedulableJobs) {
  cron.schedule(job.schedule, () => runJob(job.type, job.name));
}

// Telegram backup/health — schedule diambil dari DB settings saat startup
setTimeout(async () => {
  try {
    const { startBackupCron, startHealthCron } = await import('@/server/jobs/telegram-cron');
    await startBackupCron();
    await startHealthCron();
    console.log('[CRON] Telegram backup & health crons initialized from DB settings');
  } catch (err: any) {
    console.warn('[CRON] Telegram cron init skipped:', err?.message);
  }
}, 8_000);

console.log('[CRON RUNNER] Starting...');
console.log(`[CRON RUNNER] Loaded ${schedulableJobs.length} scheduled jobs:`);
for (const job of schedulableJobs) {
  console.log(`  ${job.scheduleLabel.padEnd(20)} → ${job.name}`);
}
console.log('[CRON RUNNER] All jobs initialized. Waiting for schedules...');

// Keep process alive
process.on('SIGTERM', () => {
  console.log('[CRON RUNNER] Received SIGTERM, shutting down gracefully...');
  process.exit(0);
});
process.on('SIGINT', () => {
  console.log('[CRON RUNNER] Received SIGINT, shutting down...');
  process.exit(0);
});
