import { NextResponse } from 'next/server';
import jwt from 'jsonwebtoken';

export async function GET(request: Request, { params }: { params: { id: string } | Promise<{ id: string }> }) {
  try {
    const resolvedParams = await params;
    const id = resolvedParams.id;
    
    const goBackendUrl = process.env.GO_API_URL || 'http://localhost:8080';
    const secret = process.env.NEXTAUTH_SECRET || '';
    const token = jwt.sign(
      { userId: 'admin', username: 'admin', role: 'admin' }, 
      secret, 
      { expiresIn: '1h' }
    );
    
    const res = await fetch(`${goBackendUrl}/api/olt/devices/${id}/onus`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
    });

    if (!res.ok) {
      return NextResponse.json({ success: false, data: [] }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json({ success: true, data: data.data || [], total: data.total || 0 });
  } catch (error: any) {
    console.error('List ONUs proxy error:', error);
    return NextResponse.json({ success: false, data: [], error: error.message }, { status: 500 });
  }
}
