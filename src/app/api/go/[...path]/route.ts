import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '@/server/auth/config';
import * as jwt from 'jsonwebtoken';

// Handle all HTTP methods
export async function GET(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}
export async function POST(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}
export async function PUT(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}
export async function DELETE(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}
export async function PATCH(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}

async function handleProxy(req: NextRequest, params: { path: string[] }) {
  try {
    const session = await getServerSession(authOptions);
    if (!session || !session.user) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    // Generate a JWT for Go backend
    const payload = {
      userId: session.user.id || 'admin',
      username: session.user.username || session.user.name || 'admin',
      name: session.user.name || 'Admin',
      role: (session.user as any).role || 'ADMIN'
    };
    
    const secret = process.env.JWT_SECRET || 'super-secret-key-change-in-production';
    const token = jwt.sign(payload, secret, { expiresIn: '1h' });

    const path = params.path.join('/');
    const searchParams = req.nextUrl.searchParams.toString();
    const targetUrl = `http://localhost:8080/api/${path}${searchParams ? '?' + searchParams : ''}`;

    const headers = new Headers();
    // Copy safe headers
    if (req.headers.has('content-type')) headers.set('content-type', req.headers.get('content-type')!);
    if (req.headers.has('accept')) headers.set('accept', req.headers.get('accept')!);
    
    // Add auth token
    headers.set('Authorization', `Bearer ${token}`);

    const init: RequestInit = {
      method: req.method,
      headers,
    };

    if (req.method !== 'GET' && req.method !== 'HEAD') {
      init.body = await req.arrayBuffer();
    }

    const response = await fetch(targetUrl, init);
    
    // Forward the response exactly as is
    const responseHeaders = new Headers(response.headers);
    responseHeaders.delete('content-encoding'); // Let fetch handle decoding
    responseHeaders.delete('content-length');
    
    return new NextResponse(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders,
    });
  } catch (error: any) {
    console.error('Go Proxy Error:', error);
    return NextResponse.json({ error: 'Proxy error', details: error.message }, { status: 500 });
  }
}
