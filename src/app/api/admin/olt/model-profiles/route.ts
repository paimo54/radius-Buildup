import { NextResponse } from 'next/server';
import { prisma } from '@/server/db/client';

export async function GET() {
  try {
    // Lakukan raw query ke tabel oid_mappings yang di-generate oleh Go GORM
    const mappings = await prisma.$queryRaw<{ vendor: string }[]>`
      SELECT DISTINCT vendor FROM oid_mappings WHERE vendor IS NOT NULL
    `;

    // Format data agar sesuai dengan yang diharapkan frontend
    const profiles = mappings.map(m => ({
      vendor: m.vendor,
      model: 'Default',
    }));

    return NextResponse.json({ profiles });
  } catch (error) {
    console.error('Error fetching vendors from DB:', error);
    return NextResponse.json({ profiles: [] });
  }
}
