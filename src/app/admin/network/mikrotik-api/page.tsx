'use client';

import { useState } from 'react';
import { Server, Activity } from 'lucide-react';
import { useTranslation } from '@/hooks/useTranslation';

export default function MikrotikApiPage() {
  const { t } = useTranslation();

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center">
          <Server className="w-5 h-5 text-cyan-500" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">API MikroTik</h1>
          <p className="text-sm text-muted-foreground">Integrasi dan manajemen API MikroTik</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="bg-card border rounded-xl p-6 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <Activity className="w-5 h-5 text-blue-500" />
            <h2 className="text-lg font-semibold">Status API</h2>
          </div>
          <p className="text-sm text-muted-foreground mb-4">
            Halaman ini sedang dalam pengembangan. Anda dapat menambahkan fitur untuk:
          </p>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-2">
            <li>Kirim perintah API ke MikroTik secara langsung</li>
            <li>Monitor resource router (CPU, Memory, Uptime) realtime</li>
            <li>Manajemen script dan scheduler MikroTik</li>
            <li>Ping dan Traceroute via MikroTik</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
