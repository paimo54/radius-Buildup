import { NextResponse } from 'next/server';
import jwt from 'jsonwebtoken';

export async function POST(request: Request, { params }: { params: { id: string } | Promise<{ id: string }> }) {
  try {
    // Await params specifically for Next.js 15+ compatibility where params is a Promise
    const resolvedParams = await params;
    const id = resolvedParams.id;
    
    // Call the Go backend for the actual SNMP scan
    const goBackendUrl = process.env.GO_API_URL || 'http://localhost:8080';
    
    // Generate a valid JWT token signed with NEXTAUTH_SECRET as expected by Go backend
    const secret = process.env.NEXTAUTH_SECRET || '';
    const token = jwt.sign(
      { userId: 'admin', username: 'admin', role: 'admin' }, 
      secret, 
      { expiresIn: '1h' }
    );
    
    const res = await fetch(`${goBackendUrl}/api/olt/devices/${id}/scan`, {
      method: 'GET', // Go backend uses GET /devices/:id/scan
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
    });

    if (!res.ok) {
      const errorText = await res.text();
      console.error(`Go backend returned ${res.status}:`, errorText);
      return NextResponse.json({ success: false, error: 'Failed to scan OLT via Go Backend. Make sure Go server is running and auth is valid.' }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json({ success: true, data: data.data || [] });
  } catch (error: any) {
    console.error('Scan proxy error:', error);
    return NextResponse.json({ success: false, error: 'Cannot connect to Go backend: ' + error.message }, { status: 500 });
  }
}
