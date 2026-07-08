import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

type ActiveTab = 'schedule' | 'employees'

interface SidebarProps {
  month: number
  year: number
  onPrevMonth: () => void
  onNextMonth: () => void
  leaveMode: 'fixed' | 'random'
  onLeaveModeChange: (mode: 'fixed' | 'random') => void
  isGenerating: boolean
  onGenerate: () => void
  onExport: () => void
  onShare: () => void
  scheduleStatus?: 'draft' | 'published'
  activeTab: ActiveTab
  onTabChange: (tab: ActiveTab) => void
}

const monthNames = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
]

export function Sidebar({
  month, year, onPrevMonth, onNextMonth,
  leaveMode, onLeaveModeChange,
  isGenerating, onGenerate, onExport, onShare,
  activeTab, onTabChange,
}: SidebarProps) {
  return (
    <aside className="w-72 bg-card border-r border-border flex flex-col shrink-0 h-screen">
      {/* Logo */}
      <div className="px-5 py-4 border-b border-border">
        <h1 className="text-xl font-bold tracking-tight">
          auto<span className="text-blue-600">Shift</span>
        </h1>
        <p className="text-xs text-muted-foreground mt-0.5">AI Shift Scheduler</p>
      </div>

      {/* Tab Navigation */}
      <div className="flex border-b border-border">
        <button
          className={cn(
            'flex-1 px-4 py-2.5 text-sm font-medium transition-colors',
            activeTab === 'schedule'
              ? 'text-primary border-b-2 border-primary'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onClick={() => onTabChange('schedule')}
        >
          Jadwal
        </button>
        <button
          className={cn(
            'flex-1 px-4 py-2.5 text-sm font-medium transition-colors',
            activeTab === 'employees'
              ? 'text-primary border-b-2 border-primary'
              : 'text-muted-foreground hover:text-foreground'
          )}
          onClick={() => onTabChange('employees')}
        >
          Karyawan
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5">
        {/* Month selector */}
        <div>
          <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Periode
          </label>
          <div className="mt-1.5 flex items-center gap-2">
            <Button variant="outline" size="icon" className="w-8 h-8" onClick={onPrevMonth}>
              ‹
            </Button>
            <span className="flex-1 text-center font-semibold">
              {monthNames[month - 1]} {year}
            </span>
            <Button variant="outline" size="icon" className="w-8 h-8" onClick={onNextMonth}>
              ›
            </Button>
          </div>
        </div>

        {/* Shift templates summary */}
        <div>
          <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Shift Aktif
          </label>
          <div className="mt-1.5 space-y-1.5">
            <ShiftBadge color="bg-blue-500" name="Pagi" time="08:00-16:00" count={5} />
            <ShiftBadge color="bg-amber-500" name="Siang" time="16:00-00:00" count={4} />
            <ShiftBadge color="bg-indigo-500" name="Malam" time="22:00-06:00" count={3} />
          </div>
        </div>

        {/* Leave mode */}
        <div>
          <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Mode Libur
          </label>
          <Select
            value={leaveMode}
            onValueChange={(v: string) => onLeaveModeChange(v as 'fixed' | 'random')}
          >
            <SelectTrigger className="mt-1.5">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="fixed">Tetap (hari tertentu)</SelectItem>
              <SelectItem value="random">Random (acak)</SelectItem>
            </SelectContent>
          </Select>
          {leaveMode === 'random' && (
            <p className="text-xs text-muted-foreground mt-1">Tanggal merah tidak mempengaruhi</p>
          )}
        </div>

        <Separator />

        {/* Generate button */}
        <Button
          className="w-full gap-2"
          onClick={onGenerate}
          disabled={isGenerating}
        >
          {isGenerating ? (
            <span className="flex items-center gap-2">
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              Generating...
            </span>
          ) : (
            <>
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              Generate Jadwal
            </>
          )}
        </Button>

        {/* Action buttons */}
        <div className="space-y-2">
          <Button variant="outline" className="w-full gap-2" onClick={onExport}>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Export PDF / Excel
          </Button>
          <Button variant="outline" className="w-full gap-2" onClick={onShare}>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            Share Link
          </Button>
        </div>
      </div>
    </aside>
  )
}

function ShiftBadge({ color, name, time, count }: {
  color: string
  name: string
  time: string
  count: number
}) {
  return (
    <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-accent/50 text-sm">
      <span className={cn('w-2 h-2 rounded-full', color)} />
      <span>{name} <span className="text-muted-foreground">({time})</span></span>
      <span className="ml-auto text-xs text-muted-foreground">{count} org</span>
    </div>
  )
}
