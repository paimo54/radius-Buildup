'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect, useCallback } from 'react';
import {
  Ticket,
  Search,
  Filter,
  RefreshCcw,
  Loader2,
  MessageSquare,
  CheckCircle2,
  Clock,
  AlertTriangle,
  Play,
  X,
  Send,
  User,
  Phone,
  ChevronDown,
  ChevronUp,
  Tag,
} from 'lucide-react';
import { useToast } from '@/components/cyberpunk/CyberToast';
import { formatWIB } from '@/lib/timezone';
import { useTranslation } from '@/hooks/useTranslation';

interface TicketData {
  id: string;
  ticketNumber: string;
  customerName: string;
  customerPhone: string;
  subject: string;
  description: string;
  priority: string;
  status: string;
  assignedToId: string | null;
  assignedToType: string | null;
  createdAt: string;
  resolvedAt: string | null;
  category: { id: string; name: string; color: string | null } | null;
  customer: { id: string; username: string; name: string; phone: string } | null;
  _count: { messages: number };
}

const PRIORITY_STYLE: Record<string, string> = {
  URGENT: 'bg-red-100 dark:bg-red-500/20 text-red-600 dark:text-red-400 border-red-300 dark:border-red-500/40',
  HIGH: 'bg-orange-100 dark:bg-orange-500/20 text-orange-600 dark:text-orange-400 border-orange-300 dark:border-orange-500/40',
  MEDIUM: 'bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 border-amber-300 dark:border-amber-500/40',
  LOW: 'bg-slate-100 dark:bg-slate-500/20 text-slate-500 dark:text-slate-400 border-slate-300 dark:border-slate-500/40',
};

const STATUS_STYLE: Record<string, string> = {
  OPEN: 'bg-amber-100 dark:bg-amber-500/20 text-amber-600 dark:text-amber-400 border-amber-300 dark:border-amber-500/40',
  IN_PROGRESS: 'bg-cyan-100 dark:bg-cyan-500/20 text-cyan-600 dark:text-[#00f7ff] border-cyan-300 dark:border-[#00f7ff]/40',
  WAITING_CUSTOMER: 'bg-purple-100 dark:bg-purple-500/20 text-purple-600 dark:text-purple-400 border-purple-300 dark:border-purple-500/40',
  RESOLVED: 'bg-green-100 dark:bg-green-500/20 text-green-600 dark:text-green-400 border-green-300 dark:border-green-500/40',
  CLOSED: 'bg-slate-100 dark:bg-slate-700/50 text-slate-500 dark:text-slate-400 border-slate-300 dark:border-slate-600/40',
};



function formatDate(d: string) {
  return formatWIB(d, 'dd MMM yyyy HH:mm');
}

export default function TechnicianTicketsPage() {
  const { t } = useTranslation();
  const { addToast } = useToast();

  const STATUS_LABEL: Record<string, string> = {
    OPEN: t('techPortal.statusOpen'),
    IN_PROGRESS: t('techPortal.statusInProgress'),
    WAITING_CUSTOMER: t('techPortal.statusWaitingCustomer'),
    RESOLVED: t('techPortal.statusResolved'),
    CLOSED: t('techPortal.statusClosed'),
  };

  const PRIORITY_LABEL: Record<string, string> = {
    URGENT: t('techPortal.priorityUrgent'),
    HIGH: t('techPortal.priorityHigh'),
    MEDIUM: t('techPortal.priorityMedium'),
    LOW: t('techPortal.priorityLow'),
  };
  const [tickets, setTickets] = useState<TicketData[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState('');
  const [filterPriority, setFilterPriority] = useState('');
  const [search, setSearch] = useState('');
  const [showMine, setShowMine] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Reply modal
  const [replyTicketId, setReplyTicketId] = useState<string | null>(null);
  const [replyMessage, setReplyMessage] = useState('');
  const [replyLoading, setReplyLoading] = useState(false);

  const loadTickets = useCallback(async () => {
    setLoading(true);
    try {
      const p = new URLSearchParams();
      if (filterStatus) p.set('status', filterStatus);
      if (filterPriority) p.set('priority', filterPriority);
      if (search) p.set('search', search);
      if (showMine) p.set('mine', 'true');
      const res = await fetch(`/api/technician/tickets?${p}`);
      if (!res.ok) throw new Error(t('techPortal.failedLoadTickets'));
      const data = await res.json();
      setTickets(data.tickets ?? []);
    } catch {
      addToast({ type: 'error', title: t('techPortal.failedLoadTickets') });
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filterStatus, filterPriority, search, showMine, addToast]);

  useEffect(() => {
    loadTickets();
  }, [loadTickets]);

  async function doAction(ticketId: string, action: string, extra?: Record<string, unknown>) {
    setActionLoading(ticketId + action);
    try {
      const res = await fetch('/api/technician/tickets', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ticketId, action, ...extra }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || t('techPortal.actionFailed'));
      addToast({ type: 'success', title: t('techPortal.actionCompleted'), description: t('techPortal.ticketUpdated') });
      loadTickets();
    } catch (e: unknown) {
      addToast({ type: 'error', title: 'Error', description: (e as Error).message });
    } finally {
      setActionLoading(null);
    }
  }

  async function sendReply() {
    if (!replyTicketId || !replyMessage.trim()) return;
    setReplyLoading(true);
    try {
      const res = await fetch('/api/technician/tickets', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ticketId: replyTicketId,
          action: 'reply',
          message: replyMessage.trim(),
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);
      addToast({ type: 'success', title: t('techPortal.replySent'), description: replyMessage.substring(0, 50) });
      setReplyTicketId(null);
      setReplyMessage('');
      loadTickets();
    } catch (e: unknown) {
      addToast({ type: 'error', title: t('techPortal.replyFailed'), description: (e as Error).message });
    } finally {
      setReplyLoading(false);
    }
  }

  const stats = {
    total: tickets.length,
    open: tickets.filter((t) => t.status === 'OPEN').length,
    inProgress: tickets.filter((t) => t.status === 'IN_PROGRESS').length,
    resolved: tickets.filter((t) => t.status === 'RESOLVED').length,
  };

  return (
    <div className="p-4 lg:p-6 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
            <Ticket className="w-5 h-5 text-[#bc13fe]" />
            {t('techPortal.tickets')}
          </h1>
          <p className="text-xs text-slate-500 dark:text-[#e0d0ff]/60 mt-0.5">
            {t('techPortal.ticketsSubtitle')}
          </p>
        </div>
        <button
          onClick={loadTickets}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-2 text-xs font-semibold bg-slate-100 dark:bg-[#bc13fe]/10 hover:bg-slate-200 dark:hover:bg-[#bc13fe]/20 text-slate-700 dark:text-[#e0d0ff] border border-slate-200 dark:border-[#bc13fe]/30 rounded-xl transition-all"
        >
          <RefreshCcw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          {t('techPortal.refresh')}
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        {[
          { label: 'Total', value: stats.total, icon: <Ticket className="w-4 h-4" />, color: 'text-[#bc13fe] bg-[#bc13fe]/10 border-[#bc13fe]/30' },
          { label: t('techPortal.statusOpen'), value: stats.open, icon: <Clock className="w-4 h-4" />, color: 'text-amber-500 bg-amber-500/10 border-amber-500/30' },
          { label: t('techPortal.statusInProgress'), value: stats.inProgress, icon: <Play className="w-4 h-4" />, color: 'text-[#00f7ff] bg-[#00f7ff]/10 border-[#00f7ff]/30' },
          { label: t('techPortal.statusResolved'), value: stats.resolved, icon: <CheckCircle2 className="w-4 h-4" />, color: 'text-green-500 bg-green-500/10 border-green-500/30' },
        ].map((s) => (
          <div key={s.label} className="bg-white dark:bg-slate-800/60 rounded-2xl border border-slate-200 dark:border-slate-700/50 p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-500 dark:text-[#e0d0ff]/60">{s.label}</p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">{s.value}</p>
              </div>
              <div className={`p-2.5 rounded-xl border ${s.color}`}>{s.icon}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="bg-white dark:bg-slate-800/60 rounded-2xl border border-slate-200 dark:border-slate-700/50 p-4">
        <div className="flex flex-wrap gap-3 items-center">
          <Filter className="w-4 h-4 text-[#00f7ff] flex-shrink-0" />
          <div className="flex-1 min-w-40 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('techPortal.searchTicket')}
              className="w-full pl-9 pr-3 py-2 text-xs bg-slate-50 dark:bg-slate-900/80 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:border-[#00f7ff]/60 focus:ring-1 focus:ring-[#00f7ff]/30"
            />
          </div>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-2 text-xs bg-slate-50 dark:bg-slate-900/80 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-[#00f7ff]/60"
          >
            <option value="">{t('techPortal.allStatus')}</option>
            <option value="OPEN">{t('techPortal.statusOpen')}</option>
            <option value="IN_PROGRESS">{t('techPortal.statusInProgress')}</option>
            <option value="WAITING_CUSTOMER">{t('techPortal.statusWaitingCustomer')}</option>
            <option value="RESOLVED">{t('techPortal.statusResolved')}</option>
            <option value="CLOSED">{t('techPortal.statusClosed')}</option>
          </select>
          <select
            value={filterPriority}
            onChange={(e) => setFilterPriority(e.target.value)}
            className="px-3 py-2 text-xs bg-slate-50 dark:bg-slate-900/80 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-[#00f7ff]/60"
          >
            <option value="">{t('techPortal.allPriority')}</option>
            <option value="URGENT">{t('techPortal.priorityUrgent')}</option>
            <option value="HIGH">{t('techPortal.priorityHigh')}</option>
            <option value="MEDIUM">{t('techPortal.priorityMedium')}</option>
            <option value="LOW">{t('techPortal.priorityLow')}</option>
          </select>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={showMine}
              onChange={(e) => setShowMine(e.target.checked)}
              className="rounded accent-[#bc13fe]"
            />
            <span className="text-xs text-slate-600 dark:text-[#e0d0ff]/70 whitespace-nowrap">{t('techPortal.myTickets')}</span>
          </label>
        </div>
      </div>

      {/* Ticket list */}
      {loading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="w-8 h-8 animate-spin text-[#00f7ff]" />
        </div>
      ) : tickets.length === 0 ? (
        <div className="text-center py-16">
          <Ticket className="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
          <p className="text-sm text-slate-500 dark:text-[#e0d0ff]/50">{t('techPortal.noTickets')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {tickets.map((ticket) => {
            const isExpanded = expandedId === ticket.id;
            const isLoading = actionLoading?.startsWith(ticket.id);
            return (
              <div
                key={ticket.id}
                className="bg-white dark:bg-slate-800/60 rounded-2xl border border-slate-200 dark:border-slate-700/50 overflow-hidden"
              >
                {/* Ticket header */}
                <div className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex flex-wrap items-center gap-2 mb-1.5">
                        <span className="text-[10px] font-mono font-bold text-[#bc13fe] dark:text-[#bc13fe]">
                          #{ticket.ticketNumber}
                        </span>
                        <span className={`px-2 py-0.5 text-[10px] font-bold rounded-full border ${PRIORITY_STYLE[ticket.priority] ?? PRIORITY_STYLE.MEDIUM}`}>
                          {PRIORITY_LABEL[ticket.priority] ?? ticket.priority}
                        </span>
                        <span className={`px-2 py-0.5 text-[10px] font-bold rounded-full border ${STATUS_STYLE[ticket.status] ?? STATUS_STYLE.OPEN}`}>
                          {STATUS_LABEL[ticket.status] ?? ticket.status}
                        </span>
                        {ticket.category && (
                          <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] rounded-full bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-600">
                            <Tag className="w-2.5 h-2.5" />
                            {ticket.category.name}
                          </span>
                        )}
                      </div>
                      <h3 className="text-sm font-semibold text-slate-900 dark:text-white truncate">
                        {ticket.subject}
                      </h3>
                      <div className="flex items-center gap-3 mt-1 text-[11px] text-slate-500 dark:text-[#e0d0ff]/60">
                        <span className="flex items-center gap-1">
                          <User className="w-3 h-3" />
                          {ticket.customerName}
                        </span>
                        <span className="flex items-center gap-1">
                          <Phone className="w-3 h-3" />
                          {ticket.customerPhone}
                        </span>
                        <span className="flex items-center gap-1">
                          <MessageSquare className="w-3 h-3" />
                          {ticket._count.messages} {t('techPortal.messages')}
                        </span>
                      </div>
                      <p className="text-[10px] text-slate-400 dark:text-slate-500 mt-0.5">
                        {formatDate(ticket.createdAt)}
                      </p>
                    </div>

                    {/* Toggle expand */}
                    <button
                      onClick={() => setExpandedId(isExpanded ? null : ticket.id)}
                      className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors flex-shrink-0"
                    >
                      {isExpanded ? (
                        <ChevronUp className="w-4 h-4 text-slate-400" />
                      ) : (
                        <ChevronDown className="w-4 h-4 text-slate-400" />
                      )}
                    </button>
                  </div>

                  {/* Action buttons */}
                  <div className="flex flex-wrap gap-2 mt-3 pt-3 border-t border-slate-100 dark:border-slate-700/50">
                    {/* Claim - show if unassigned or not mine */}
                    {(!ticket.assignedToId) && ticket.status !== 'RESOLVED' && ticket.status !== 'CLOSED' && (
                      <button
                        onClick={() => doAction(ticket.id, 'claim')}
                        disabled={!!isLoading}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-semibold bg-[#bc13fe]/10 hover:bg-[#bc13fe]/20 text-[#bc13fe] border border-[#bc13fe]/30 rounded-xl transition-all"
                      >
                        {isLoading ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3" />}
                        {t('techPortal.claimTicket')}
                      </button>
                    )}

                    {/* Status updates */}
                    {ticket.status === 'IN_PROGRESS' && (
                      <>
                        <button
                          onClick={() => doAction(ticket.id, 'update_status', { status: 'WAITING_CUSTOMER' })}
                          disabled={!!isLoading}
                          className="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-semibold bg-purple-50 dark:bg-purple-500/10 hover:bg-purple-100 dark:hover:bg-purple-500/20 text-purple-600 dark:text-purple-400 border border-purple-200 dark:border-purple-500/30 rounded-xl transition-all"
                        >
                          <Clock className="w-3 h-3" />
                          {t('techPortal.waitingCustomer')}
                        </button>
                        <button
                          onClick={() => doAction(ticket.id, 'update_status', { status: 'RESOLVED' })}
                          disabled={!!isLoading}
                          className="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-semibold bg-green-50 dark:bg-green-500/10 hover:bg-green-100 dark:hover:bg-green-500/20 text-green-600 dark:text-green-400 border border-green-200 dark:border-green-500/30 rounded-xl transition-all"
                        >
                          <CheckCircle2 className="w-3 h-3" />
                          {t('techPortal.markDone')}
                        </button>
                      </>
                    )}
                    {ticket.status === 'WAITING_CUSTOMER' && (
                      <button
                        onClick={() => doAction(ticket.id, 'update_status', { status: 'IN_PROGRESS' })}
                        disabled={!!isLoading}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-semibold bg-cyan-50 dark:bg-cyan-500/10 hover:bg-cyan-100 dark:hover:bg-cyan-500/20 text-cyan-600 dark:text-[#00f7ff] border border-cyan-200 dark:border-[#00f7ff]/30 rounded-xl transition-all"
                      >
                        <Play className="w-3 h-3" />
                        Lanjutkan
                      </button>
                    )}

                    {/* Reply */}
                    {ticket.status !== 'CLOSED' && (
                      <button
                        onClick={() => { setReplyTicketId(ticket.id); setReplyMessage(''); }}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-semibold bg-slate-100 dark:bg-slate-700/50 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-xl transition-all"
                      >
                        <MessageSquare className="w-3 h-3" />
                        {t('techPortal.replyTicket')}
                      </button>
                    )}
                  </div>
                </div>

                {/* Expanded description */}
                {isExpanded && (
                  <div className="px-4 pb-4 bg-slate-50/50 dark:bg-slate-900/30 border-t border-slate-100 dark:border-slate-700/50">
                    <p className="text-xs text-slate-700 dark:text-[#e0d0ff]/80 mt-3 whitespace-pre-wrap leading-relaxed">
                      {ticket.description}
                    </p>
                    {ticket.customer && (
                      <div className="mt-3 flex items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400">
                        <span>Username RADIUS:</span>
                        <span className="font-mono font-semibold text-[#00f7ff]">{ticket.customer.username}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Reply modal */}
      {replyTicketId && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-end sm:items-center justify-center p-4">
          <div className="w-full max-w-lg bg-white dark:bg-[#0a0520] rounded-2xl border border-slate-200 dark:border-[#bc13fe]/30 shadow-2xl dark:shadow-[0_0_40px_rgba(188,19,254,0.2)]">
            <div className="flex items-center justify-between p-4 border-b border-slate-100 dark:border-[#bc13fe]/20">
              <h3 className="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <MessageSquare className="w-4 h-4 text-[#bc13fe]" />
                {t('techPortal.replyTitle')}
              </h3>
              <button
                onClick={() => setReplyTicketId(null)}
                className="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg"
              >
                <X className="w-4 h-4 text-slate-400" />
              </button>
            </div>
            <div className="p-4 space-y-3">
              <textarea
                value={replyMessage}
                onChange={(e) => setReplyMessage(e.target.value)}
                rows={5}
                placeholder={t('techPortal.replyPlaceholder')}
                className="w-full px-3 py-2.5 text-sm bg-slate-50 dark:bg-slate-900/80 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:border-[#00f7ff]/60 focus:ring-1 focus:ring-[#00f7ff]/30 resize-none"
              />
              <div className="flex gap-2">
                <button
                  onClick={() => setReplyTicketId(null)}
                  className="flex-1 py-2.5 text-xs font-semibold bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 rounded-xl border border-slate-200 dark:border-slate-600 hover:bg-slate-200 dark:hover:bg-slate-600 transition-all"
                >
                  {t('techPortal.cancel')}
                </button>
                <button
                  onClick={sendReply}
                  disabled={replyLoading || !replyMessage.trim()}
                  className="flex-1 flex items-center justify-center gap-2 py-2.5 text-xs font-bold bg-[#bc13fe] hover:bg-[#bc13fe]/90 text-white rounded-xl transition-all disabled:opacity-50"
                >
                  {replyLoading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
                  {t('techPortal.sendReply')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
