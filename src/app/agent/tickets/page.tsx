'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import {
  TicketPlus,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Send,
  Tag,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  MessageSquare,
  Plus,
} from 'lucide-react';
import { showSuccess, showError, showConfirm } from '@/lib/sweetalert';
import { useTranslation } from '@/hooks/useTranslation';
import { formatWIB } from '@/lib/timezone';

interface TicketMessage {
  id: string;
  senderType: string;
  senderName: string;
  message: string;
  createdAt: string;
}

interface Ticket {
  id: string;
  ticketNumber: string;
  subject: string;
  description: string;
  priority: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  category?: { name: string; color?: string } | null;
  messages: TicketMessage[];
  _count?: { messages: number };
}

interface Category {
  id: string;
  name: string;
  color?: string;
}

const STATUS_COLORS: Record<string, string> = {
  OPEN:             'bg-blue-500/10 text-blue-400 border-blue-500/30',
  IN_PROGRESS:      'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
  WAITING_CUSTOMER: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
  RESOLVED:         'bg-green-500/10 text-green-400 border-green-500/30',
  CLOSED:           'bg-slate-500/10 text-slate-400 border-slate-500/30',
};

const PRIORITY_COLORS: Record<string, string> = {
  LOW:    'bg-slate-500/10 text-slate-400 border-slate-500/30',
  MEDIUM: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
  HIGH:   'bg-orange-500/10 text-orange-400 border-orange-500/30',
  URGENT: 'bg-red-500/10 text-red-400 border-red-500/30',
};

const STATUS_ICON: Record<string, React.ReactNode> = {
  OPEN:             <AlertCircle className="w-3.5 h-3.5" />,
  IN_PROGRESS:      <Clock className="w-3.5 h-3.5" />,
  WAITING_CUSTOMER: <Clock className="w-3.5 h-3.5" />,
  RESOLVED:         <CheckCircle2 className="w-3.5 h-3.5" />,
  CLOSED:           <XCircle className="w-3.5 h-3.5" />,
};

function fmtDate(val: string) {
  return formatWIB(val, 'dd MMM yyyy HH:mm');
}

export default function AgentTicketsPage() {
  const router = useRouter();
  const { t } = useTranslation();

  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState('');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [replyText, setReplyText] = useState<Record<string, string>>({});
  const [sendingReply, setSendingReply] = useState<string | null>(null);

  // New ticket form
  const [showForm, setShowForm] = useState(false);
  const [creating, setCreating] = useState(false);
  const [form, setForm] = useState({ subject: '', description: '', priority: 'MEDIUM', categoryId: '' });

  const messagesRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const data = localStorage.getItem('agentData');
    if (!data) { router.push('/agent'); return; }
    loadTickets();
    loadCategories();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router]);

  const loadTickets = async (status = filterStatus) => {
    setLoading(true);
    try {
      const token = localStorage.getItem('agentToken');
      if (!token) { router.push('/agent'); return; }
      const params = new URLSearchParams();
      if (status) params.append('status', status);
      const res = await fetch(`/api/agent/tickets?${params}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.success) setTickets(data.tickets);
    } catch {
      // silently ignore
    } finally {
      setLoading(false);
    }
  };

  const loadCategories = async () => {
    try {
      const res = await fetch('/api/tickets/categories');
      if (res.ok) {
        const data = await res.json();
        setCategories(Array.isArray(data) ? data : []);
      }
    } catch { /* ignore */ }
  };

  const handleCreate = async () => {
    if (!form.subject.trim() || !form.description.trim()) {
      await showError('Subjek dan deskripsi wajib diisi.');
      return;
    }
    setCreating(true);
    try {
      const token = localStorage.getItem('agentToken');
      const res = await fetch('/api/agent/tickets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ ...form }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        await showSuccess(`Tiket #${data.ticket.ticketNumber} berhasil dibuat!`);
        setForm({ subject: '', description: '', priority: 'MEDIUM', categoryId: '' });
        setShowForm(false);
        loadTickets();
      } else {
        await showError(data.error || 'Gagal membuat tiket.');
      }
    } catch {
      await showError('Gagal membuat tiket.');
    } finally {
      setCreating(false);
    }
  };

  const handleReply = async (ticketId: string) => {
    const msg = replyText[ticketId]?.trim();
    if (!msg) return;
    setSendingReply(ticketId);
    try {
      const token = localStorage.getItem('agentToken');
      const res = await fetch(`/api/agent/tickets/${ticketId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ message: msg }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        setReplyText(prev => ({ ...prev, [ticketId]: '' }));
        // Reload tickets to refresh messages
        await loadTickets(filterStatus);
        // Keep expanded
        setExpandedId(ticketId);
      } else {
        await showError(data.error || 'Gagal mengirim balasan.');
      }
    } catch {
      await showError('Gagal mengirim balasan.');
    } finally {
      setSendingReply(null);
    }
  };

  const handleFilterChange = (status: string) => {
    setFilterStatus(status);
    loadTickets(status);
  };

  const statusLabel: Record<string, string> = {
    '': 'Semua',
    OPEN: 'Terbuka',
    IN_PROGRESS: 'Diproses',
    WAITING_CUSTOMER: 'Menunggu',
    RESOLVED: 'Selesai',
    CLOSED: 'Ditutup',
  };
  const priorityLabel: Record<string, string> = {
    LOW: 'Rendah', MEDIUM: 'Sedang', HIGH: 'Tinggi', URGENT: 'Mendesak',
  };

  return (
    <div className="p-4 sm:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-black tracking-wide bg-gradient-to-r from-[#00f7ff] to-[#bc13fe] bg-clip-text text-transparent">
            Tiket Keluhan / Gangguan
          </h1>
          <p className="text-xs text-slate-500 dark:text-[#e0d0ff]/60 mt-0.5">
            Laporkan gangguan atau keluhan Anda kepada tim support
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => loadTickets()}
            className="flex items-center gap-1.5 px-3 py-2 text-xs font-semibold rounded-xl bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-white/10 transition-all"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            Refresh
          </button>
          <button
            onClick={() => setShowForm(true)}
            className="flex items-center gap-1.5 px-4 py-2 text-xs font-bold rounded-xl bg-gradient-to-r from-[#bc13fe] to-[#00f7ff] text-white shadow-[0_0_15px_rgba(188,19,254,0.4)] hover:shadow-[0_0_25px_rgba(0,247,255,0.4)] transition-all"
          >
            <Plus className="w-3.5 h-3.5" />
            Buat Tiket
          </button>
        </div>
      </div>

      {/* Create Ticket Form */}
      {showForm && (
        <div className="rounded-2xl bg-white dark:bg-[#0a0520]/80 border border-purple-200 dark:border-[#bc13fe]/30 shadow-[0_0_30px_rgba(188,19,254,0.1)] p-5 space-y-4">
          <h2 className="text-sm font-bold text-slate-800 dark:text-white flex items-center gap-2">
            <TicketPlus className="w-4 h-4 text-[#bc13fe]" />
            Buat Tiket Baru
          </h2>

          {/* Subject */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-slate-600 dark:text-[#e0d0ff]/80">Subjek <span className="text-red-400">*</span></label>
            <input
              value={form.subject}
              onChange={e => setForm(f => ({ ...f, subject: e.target.value }))}
              placeholder="Contoh: Internet mati sejak pagi"
              className="w-full px-3 py-2 text-sm rounded-xl bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-[#bc13fe]/20 text-slate-800 dark:text-white placeholder:text-slate-400 dark:placeholder:text-white/30 focus:outline-none focus:border-[#00f7ff]/50 transition"
            />
          </div>

          {/* Description */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-slate-600 dark:text-[#e0d0ff]/80">Deskripsi <span className="text-red-400">*</span></label>
            <textarea
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              placeholder="Jelaskan masalah secara detail..."
              rows={4}
              className="w-full px-3 py-2 text-sm rounded-xl bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-[#bc13fe]/20 text-slate-800 dark:text-white placeholder:text-slate-400 dark:placeholder:text-white/30 focus:outline-none focus:border-[#00f7ff]/50 transition resize-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            {/* Priority */}
            <div className="space-y-1">
              <label className="text-xs font-semibold text-slate-600 dark:text-[#e0d0ff]/80">Prioritas</label>
              <select
                value={form.priority}
                onChange={e => setForm(f => ({ ...f, priority: e.target.value }))}
                className="w-full px-3 py-2 text-sm rounded-xl bg-slate-50 dark:bg-[#0d0a1e] border border-slate-200 dark:border-[#bc13fe]/30 text-slate-800 dark:text-white focus:outline-none focus:border-[#bc13fe]/60 transition appearance-none cursor-pointer"
              >
                {Object.entries(priorityLabel).map(([val, label]) => (
                  <option key={val} value={val} className="bg-background dark:bg-[#0d0a1e] text-foreground dark:text-white">{label}</option>
                ))}
              </select>
            </div>

            {/* Category */}
            <div className="space-y-1">
              <label className="text-xs font-semibold text-slate-600 dark:text-[#e0d0ff]/80">Kategori</label>
              <select
                value={form.categoryId}
                onChange={e => setForm(f => ({ ...f, categoryId: e.target.value }))}
                className="w-full px-3 py-2 text-sm rounded-xl bg-slate-50 dark:bg-[#0d0a1e] border border-slate-200 dark:border-[#bc13fe]/30 text-slate-800 dark:text-white focus:outline-none focus:border-[#bc13fe]/60 transition appearance-none cursor-pointer"
              >
                <option value="" className="bg-background dark:bg-[#0d0a1e] text-slate-400">-- Pilih Kategori --</option>
                {categories.map(c => (
                  <option key={c.id} value={c.id} className="bg-background dark:bg-[#0d0a1e] text-foreground dark:text-white">{c.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex gap-2 pt-1">
            <button
              onClick={() => { setShowForm(false); setForm({ subject: '', description: '', priority: 'MEDIUM', categoryId: '' }); }}
              className="flex-1 py-2 text-xs font-bold rounded-xl bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-white/10 transition"
            >
              Batal
            </button>
            <button
              onClick={handleCreate}
              disabled={creating}
              className="flex-1 flex items-center justify-center gap-2 py-2 text-xs font-bold rounded-xl bg-gradient-to-r from-[#bc13fe] to-[#00f7ff] text-white disabled:opacity-60 shadow-[0_0_15px_rgba(188,19,254,0.3)] transition"
            >
              {creating ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
              {creating ? 'Mengirim...' : 'Kirim Tiket'}
            </button>
          </div>
        </div>
      )}

      {/* Filter */}
      <div className="flex gap-2 flex-wrap">
        {Object.entries(statusLabel).map(([val, label]) => (
          <button
            key={val}
            onClick={() => handleFilterChange(val)}
            className={`px-3 py-1.5 text-xs font-semibold rounded-xl border transition-all ${
              filterStatus === val
                ? 'bg-[#bc13fe]/20 text-[#bc13fe] border-[#bc13fe]/40 shadow-[0_0_10px_rgba(188,19,254,0.2)]'
                : 'bg-slate-100 dark:bg-white/5 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-white/10 hover:bg-slate-200 dark:hover:bg-white/10'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Ticket List */}
      {loading ? (
        <div className="flex justify-center items-center py-16">
          <Loader2 className="w-6 h-6 text-[#bc13fe] animate-spin" />
        </div>
      ) : tickets.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3 text-center">
          <div className="p-5 rounded-2xl bg-[#bc13fe]/10 border border-[#bc13fe]/20">
            <MessageSquare className="w-10 h-10 text-[#bc13fe]/60" />
          </div>
          <p className="text-sm font-bold text-slate-600 dark:text-slate-300">Belum ada tiket</p>
          <p className="text-xs text-slate-400 dark:text-slate-500">Buat tiket baru untuk melaporkan gangguan</p>
        </div>
      ) : (
        <div className="space-y-3">
          {tickets.map(ticket => (
            <div
              key={ticket.id}
              className="rounded-2xl bg-white dark:bg-[#0a0520]/80 border border-slate-200 dark:border-[#bc13fe]/20 shadow-sm dark:shadow-[0_0_15px_rgba(188,19,254,0.05)] overflow-hidden transition-all"
            >
              {/* Ticket Header */}
              <button
                onClick={() => setExpandedId(expandedId === ticket.id ? null : ticket.id)}
                className="w-full text-left px-4 py-3.5 flex items-start justify-between gap-3 hover:bg-slate-50 dark:hover:bg-white/5 transition"
              >
                <div className="flex-1 min-w-0 space-y-1.5">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-[10px] font-bold text-[#bc13fe] bg-[#bc13fe]/10 px-2 py-0.5 rounded-md">
                      #{ticket.ticketNumber}
                    </span>
                    <span className={`flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-md border ${STATUS_COLORS[ticket.status] || STATUS_COLORS.OPEN}`}>
                      {STATUS_ICON[ticket.status]}
                      {ticket.status.replace('_', ' ')}
                    </span>
                    <span className={`text-[10px] font-bold px-2 py-0.5 rounded-md border ${PRIORITY_COLORS[ticket.priority] || PRIORITY_COLORS.MEDIUM}`}>
                      {priorityLabel[ticket.priority] || ticket.priority}
                    </span>
                    {ticket.category && (
                      <span className="flex items-center gap-1 text-[10px] font-semibold text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-white/5 px-2 py-0.5 rounded-md">
                        <Tag className="w-3 h-3" />
                        {ticket.category.name}
                      </span>
                    )}
                  </div>
                  <p className="text-sm font-bold text-slate-800 dark:text-white truncate">{ticket.subject}</p>
                  <div className="flex items-center gap-3 text-[10px] text-slate-400 dark:text-slate-500">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {fmtDate(ticket.createdAt)}
                    </span>
                    <span className="flex items-center gap-1">
                      <MessageSquare className="w-3 h-3" />
                      {ticket._count?.messages || ticket.messages.length} pesan
                    </span>
                  </div>
                </div>
                <span className="flex-shrink-0 text-slate-400 mt-1">
                  {expandedId === ticket.id ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                </span>
              </button>

              {/* Expanded - Messages + Reply */}
              {expandedId === ticket.id && (
                <div className="border-t border-slate-100 dark:border-[#bc13fe]/10">
                  {/* Messages */}
                  <div className="px-4 py-3 space-y-3 max-h-72 overflow-y-auto" ref={messagesRef}>
                    {ticket.messages.map(msg => {
                      const isAgent = msg.senderType === 'CUSTOMER';
                      return (
                        <div key={msg.id} className={`flex gap-2.5 ${isAgent ? 'flex-row-reverse' : 'flex-row'}`}>
                          <div className={`flex-shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-[10px] font-black ${
                            isAgent
                              ? 'bg-gradient-to-br from-[#bc13fe] to-[#00f7ff] text-white'
                              : 'bg-slate-200 dark:bg-white/10 text-slate-600 dark:text-white'
                          }`}>
                            {msg.senderName.charAt(0).toUpperCase()}
                          </div>
                          <div className={`flex-1 max-w-[80%] space-y-0.5 ${isAgent ? 'items-end' : 'items-start'} flex flex-col`}>
                            <div className={`px-3 py-2 rounded-xl text-xs leading-relaxed ${
                              isAgent
                                ? 'bg-gradient-to-br from-[#bc13fe]/20 to-[#00f7ff]/10 dark:from-[#bc13fe]/30 dark:to-[#00f7ff]/20 text-slate-800 dark:text-white border border-[#bc13fe]/20'
                                : 'bg-slate-100 dark:bg-white/10 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-white/10'
                            }`}>
                              {msg.message}
                            </div>
                            <div className={`text-[9px] text-slate-400 dark:text-slate-500 px-1 ${isAgent ? 'text-right' : 'text-left'}`}>
                              {msg.senderType === 'ADMIN' ? '👤 Admin' : msg.senderType === 'TECHNICIAN' ? '🔧 Teknisi' : msg.senderName} · {fmtDate(msg.createdAt)}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  {/* Reply box - only if not closed */}
                  {ticket.status !== 'CLOSED' && (
                    <div className="px-4 pb-4 flex gap-2">
                      <input
                        value={replyText[ticket.id] || ''}
                        onChange={e => setReplyText(prev => ({ ...prev, [ticket.id]: e.target.value }))}
                        onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleReply(ticket.id); } }}
                        placeholder="Tulis balasan..."
                        className="flex-1 px-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-[#bc13fe]/20 text-slate-800 dark:text-white placeholder:text-slate-400 dark:placeholder:text-white/30 focus:outline-none focus:border-[#00f7ff]/50 transition"
                      />
                      <button
                        onClick={() => handleReply(ticket.id)}
                        disabled={sendingReply === ticket.id || !replyText[ticket.id]?.trim()}
                        className="flex items-center gap-1 px-3 py-2 text-xs font-bold rounded-xl bg-gradient-to-r from-[#bc13fe] to-[#00f7ff] text-white disabled:opacity-50 transition"
                      >
                        {sendingReply === ticket.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
