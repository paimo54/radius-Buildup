import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'
import { authOptions } from '@/server/auth/config'
import { MikroTikConnection } from '@/server/services/mikrotik/client'

// POST - Test router connection
export async function POST(request: Request) {
  const session = await getServerSession(authOptions)
  
  if (!session) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  try {
    const { ipAddress, username, password, port } = await request.json()

    if (!ipAddress || !username || !password) {
      return NextResponse.json(
        { error: 'IP Address, Username, and Password are required' },
        { status: 400 }
      )
    }

    const mtik = new MikroTikConnection({
      host: ipAddress,
      username,
      password,
      port: parseInt(port) || 8728,
      timeout: 15000,
    })

    const result = await mtik.testConnection()

    return NextResponse.json(result)
  } catch (error: any) {
    console.error('Test router connection error:', error)
    return NextResponse.json({
      success: false,
      message: error.message || 'Connection test failed',
    })
  }
}
