'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/useTranslation';
import { ArrowLeft, Server, Activity, RefreshCw, AlertTriangle } from 'lucide-react';
import { showSuccess, showError } from '@/lib/sweetalert';
import Link from 'next/link';

export default function OltDetailPage() {
  const { t } = useTranslation();
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [olt, setOlt] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [onus, setOnus] = useState<any[]>([]);

  useEffect(() => {
    if (id) {
      loadOltDetails();
      loadCachedOnus(); // Load cached data from DB on page load
    }
  }, [id]);

  const loadOltDetails = async () => {
    try {
      const res = await fetch('/api/network/olts');
      const data = await res.json();
      const found = data.olts?.find((o: any) => o.id === id);
      
      if (found) {
        setOlt(found);
      } else {
        showError('Error', 'OLT not found');
        router.push('/admin/network/olts');
      }
    } catch (error) {
      console.error('Failed to load OLT:', error);
      showError('Error', 'Failed to load OLT details');
    } finally {
      setLoading(false);
    }
  };

  // Load cached ONU data from database (last scan results)
  const loadCachedOnus = async () => {
    try {
      const res = await fetch(`/api/admin/olt/${id}/onus`);
      const data = await res.json();
      if (data.data && data.data.length > 0) {
        setOnus(data.data);
        console.log(`[OLT] Loaded ${data.data.length} cached ONUs from database`);
      }
    } catch (error) {
      console.error('Failed to load cached ONUs:', error);
    }
  };

  const handleScan = async () => {
    setScanning(true);
    try {
      // Use AbortController to extend timeout (SNMP scan can take 10-30s for large OLTs)
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 60000); // 60 second timeout
      
      const res = await fetch(`/api/admin/olt/${id}/scan`, {
        method: 'POST',
        signal: controller.signal,
      });
      clearTimeout(timeoutId);
      
      const data = await res.json();
      console.log('[OLT Scan] Response:', { success: data.success, total: data.data?.length, error: data.error });
      
      if (data.success && data.data) {
        setOnus(data.data);
        showSuccess('Scan Berhasil', `Ditemukan ${data.data.length} ONU`);
      } else if (data.data && Array.isArray(data.data)) {
        setOnus(data.data);
        showSuccess('Scan Berhasil', `Ditemukan ${data.data.length} ONU`);
      } else {
        showError('Scan Gagal', data.error || 'Response tidak valid');
      }
    } catch (error: any) {
      console.error('Failed to scan:', error);
      if (error.name === 'AbortError') {
        showError('Scan Timeout', 'Scan OLT membutuhkan waktu terlalu lama. Coba lagi.');
      } else {
        showError('Scan Gagal', 'Network error: ' + error.message);
      }
    } finally {
      setScanning(false);
    }
  };

  const stats = {
    total: onus.length,
    online: onus.filter(o => {
      const s = o.status?.toLowerCase() || '';
      return s === 'online' || s === 'up';
    }).length,
    offline: onus.filter(o => {
      const s = o.status?.toLowerCase() || '';
      return s !== 'online' && s !== 'up';
    }).length,
    weakSignal: onus.filter(o => {
      const s = o.status?.toLowerCase() || '';
      const isOnline = s === 'online' || s === 'up';
      // Signal is weak if Rx is less than -25 dBm. Note: if Rx is perfectly 0 or exactly 12.90, we ignore it as anomaly for weak signal
      return isOnline && o.rx !== undefined && o.rx !== 0 && o.rx < -25;
    }).length,
  };

  const getRxColor = (rx: number) => {
    if (rx < -27) return 'text-red-600 font-bold';
    if (rx < -25) return 'text-orange-500 font-bold';
    if (rx >= -25 && rx < 0) return 'text-green-600 font-medium';
    return 'text-gray-700 dark:text-gray-300';
  };

  if (loading) {
    return <div className="p-6 text-center text-gray-500">Loading...</div>;
  }

  if (!olt) return null;

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/admin/network/olts" className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Server className="h-6 w-6 text-teal-600" />
              {olt.name}
            </h1>
            <p className="text-sm text-gray-500">IP: {olt.ipAddress} • Vendor: {olt.vendor?.toUpperCase()}</p>
          </div>
        </div>
        <button
          onClick={handleScan}
          disabled={scanning}
          className="flex items-center gap-2 px-4 py-2 bg-teal-600 hover:bg-teal-700 text-white rounded-lg transition-colors disabled:opacity-50 shadow-sm"
        >
          <RefreshCw className={`h-4 w-4 ${scanning ? 'animate-spin' : ''}`} />
          {scanning ? 'Menarik Data SNMP...' : 'Sync Data OLT'}
        </button>
      </div>

      {/* Statistic Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-900 p-4 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800 flex items-center gap-4 cursor-pointer hover:border-blue-500 transition-colors">
          <div className="p-3 bg-blue-50 dark:bg-blue-900/30 text-blue-600 rounded-lg">
            <Activity className="h-6 w-6" />
          </div>
          <div>
            <p className="text-2xl font-bold">{stats.total}</p>
            <p className="text-xs text-gray-500 uppercase tracking-wide">Total ONU</p>
          </div>
        </div>
        <div className="bg-white dark:bg-gray-900 p-4 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800 flex items-center gap-4 cursor-pointer hover:border-green-500 transition-colors">
          <div className="p-3 bg-green-50 dark:bg-green-900/30 text-green-600 rounded-lg">
            <Activity className="h-6 w-6" />
          </div>
          <div>
            <p className="text-2xl font-bold">{stats.online}</p>
            <p className="text-xs text-gray-500 uppercase tracking-wide">ONU Online</p>
          </div>
        </div>
        <div className="bg-white dark:bg-gray-900 p-4 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800 flex items-center gap-4 cursor-pointer hover:border-red-500 transition-colors">
          <div className="p-3 bg-red-50 dark:bg-red-900/30 text-red-600 rounded-lg">
            <Activity className="h-6 w-6" />
          </div>
          <div>
            <p className="text-2xl font-bold">{stats.offline}</p>
            <p className="text-xs text-gray-500 uppercase tracking-wide">ONU Offline</p>
          </div>
        </div>
        <div className="bg-white dark:bg-gray-900 p-4 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800 flex items-center gap-4 cursor-pointer hover:border-orange-500 transition-colors">
          <div className="p-3 bg-orange-50 dark:bg-orange-900/30 text-orange-600 rounded-lg">
            <AlertTriangle className="h-6 w-6" />
          </div>
          <div>
            <p className="text-2xl font-bold">{stats.weakSignal}</p>
            <p className="text-xs text-gray-500 uppercase tracking-wide">Sinyal Lemah</p>
          </div>
        </div>
      </div>

      {/* Scanned ONUs Table */}
      <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800 overflow-hidden">
        <div className="p-5 border-b border-gray-100 dark:border-gray-800 flex justify-between items-center">
          <h3 className="text-base font-semibold flex items-center gap-2">
            <Activity className="h-5 w-5 text-gray-500" />
            Monitoring ONU (SNMP)
          </h3>
        </div>
        
        {onus.length === 0 ? (
          <div className="p-12 text-center text-gray-500 flex flex-col items-center">
            <AlertTriangle className="h-12 w-12 text-gray-300 mb-3" />
            <p>Data kosong. Silakan klik Sync Data OLT.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-gray-50 dark:bg-gray-800/50 text-gray-500 text-xs uppercase">
                <tr>
                  <th className="px-4 py-3">Index</th>
                  <th className="px-4 py-3">User / Nama ONU</th>
                  <th className="px-4 py-3">SN / Serial</th>
                  <th className="px-4 py-3">Model</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Tx (OLT Rx)</th>
                  <th className="px-4 py-3">Rx (ONU Rx)</th>
                  <th className="px-4 py-3">Jarak</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                {onus.map((onu, i) => (
                  <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                    <td className="px-4 py-3 font-mono text-xs text-gray-500">{onu.index}</td>
                    <td className="px-4 py-3 font-medium">{onu.name || '-'}</td>
                    <td className="px-4 py-3 font-mono text-[10px]">{onu.serial || '-'}</td>
                    <td className="px-4 py-3 text-xs">{onu.model || '-'}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 text-[10px] rounded font-medium ${(onu.status === 'Online' || onu.status === 'Up') ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'}`}>
                        {(onu.status === 'Online' || onu.status === 'Up') ? '● Online' : '○ Offline'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs">
                      {onu.tx !== undefined && onu.tx !== 0 && onu.tx > -40 ? `${onu.tx.toFixed(2)} dBm` : 'N/A'}
                    </td>
                    <td className="px-4 py-3 text-xs">
                      {onu.rx !== undefined && onu.rx !== 0 && onu.rx > -40 ? (
                        <span className={getRxColor(onu.rx)}>{onu.rx.toFixed(2)} dBm</span>
                      ) : 'N/A'}
                    </td>
                    <td className="px-4 py-3 text-xs">
                      {onu.distance ? `${onu.distance} m` : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
