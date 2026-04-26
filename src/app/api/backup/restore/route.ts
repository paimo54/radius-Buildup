import { NextRequest, NextResponse } from 'next/server';
import { restoreBackup } from '@/server/services/backup.service';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/server/auth/config';
import { writeFile, unlink } from 'fs/promises';
import path from 'path';

// Allow up to 5 minutes for large restore operations
export const maxDuration = 300;

export async function POST(request: NextRequest) {
  try {
    // Check authentication
    const session = await getServerSession(authOptions);
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    // Check if user is SUPER_ADMIN
    if (session.user.role !== 'SUPER_ADMIN') {
      return NextResponse.json({ error: 'Forbidden: Only SUPER_ADMIN can restore database' }, { status: 403 });
    }

    console.log(`[Restore API] User ${session.user.username} initiated database restore`);

    const formData = await request.formData();
    const file = formData.get('file') as File;

    if (!file) {
      return NextResponse.json({ error: 'No file provided' }, { status: 400 });
    }

    const isGzip = file.name.endsWith('.sql.gz') || file.name.endsWith('.gz');
    const isSql = file.name.endsWith('.sql');

    if (!isSql && !isGzip) {
      return NextResponse.json({ error: 'File must be .sql or .sql.gz format' }, { status: 400 });
    }

    // Save uploaded file temporarily
    const bytes = await file.arrayBuffer();
    const buffer = Buffer.from(bytes);

    const ext = isGzip ? (file.name.endsWith('.sql.gz') ? '.sql.gz' : '.gz') : '.sql';
    const tempFilepath = path.join(process.cwd(), 'backups', `restore_temp_${Date.now()}${ext}`);
    await writeFile(tempFilepath, buffer);

    console.log('[Restore API] File uploaded, starting restore...');

    try {
      // restoreBackup handles both .sql and .sql.gz/.gz
      await restoreBackup(tempFilepath);
    } finally {
      // Always clean up temp file
      await unlink(tempFilepath).catch(() => {});
    }

    console.log('[Restore API] Database restored successfully');

    return NextResponse.json({
      success: true,
      message: 'Database restored successfully',
    });
  } catch (error: any) {
    console.error('[Restore API] Error:', error);
    return NextResponse.json(
      { success: false, error: error.message },
      { status: 500 }
    );
  }
}
