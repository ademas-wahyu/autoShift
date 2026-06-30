import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Sidebar } from '@/components/sidebar/Sidebar'
import { CalendarGrid } from '@/components/calendar/CalendarGrid'
import type {
  ScheduleShift, ScheduleLeave, CalendarCell,
  Employee, ValidationViolation,
} from '@/types'

// ─── Mock data for prototype ───────────────────────────────
const MOCK_EMPLOYEES: Employee[] = [
  { id: 1, name: 'Budi', role_id: 1, role_name: 'Supervisor', is_active: true },
  { id: 2, name: 'Siti', role_id: 2, role_name: 'Staff', is_active: true },
  { id: 3, name: 'Andi', role_id: 2, role_name: 'Staff', is_active: true },
  { id: 4, name: 'Dewi', role_id: 2, role_name: 'Staff', is_active: true },
  { id: 5, name: 'Rudi', role_id: 1, role_name: 'Supervisor', is_active: true },
]

function generateWeekDays(year: number, month: number) {
  const firstDay = new Date(year, month - 1, 1)
  const start = new Date(firstDay)
  start.setDate(start.getDate() - start.getDay()) // go to Sunday

  const days: { label: string; date: string }[] = []
  const labels = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']
  for (let i = 0; i < 35; i++) {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    days.push({
      label: labels[d.getDay()],
      date: d.toISOString().split('T')[0],
    })
  }
  return days
}

export default function App() {
  const now = new Date()
  const [month, setMonth] = useState(now.getMonth() + 1)
  const [year, setYear] = useState(now.getFullYear())
  const [leaveMode, setLeaveMode] = useState<'fixed' | 'random'>('random')
  const [isGenerating, setIsGenerating] = useState(false)
  const [scheduleStatus, setScheduleStatus] = useState<'draft' | 'published' | null>(null)

  // Mock state
  const [shifts, setShifts] = useState<ScheduleShift[]>([])
  const [leaves, setLeaves] = useState<ScheduleLeave[]>([])
  const [violations, setViolations] = useState<ValidationViolation[]>([])

  const weekDays = useMemo(() => generateWeekDays(year, month), [month, year])

  // Generate calendar cells metadata
  const cells: CalendarCell[] = useMemo(() =>
    weekDays.map((d) => {
      const dateObj = new Date(d.date + 'T00:00:00')
      return {
        date: d.date,
        dayOfWeek: dateObj.getDay(),
        isHoliday: d.label === 'Min',
        holidayName: d.label === 'Min' ? 'Minggu' : undefined,
        isToday: d.date === now.toISOString().split('T')[0],
        isWeekend: dateObj.getDay() === 0 || dateObj.getDay() === 6,
      }
    }), [weekDays])

  const handleGenerate = () => {
    setIsGenerating(true)
    setScheduleStatus('draft')
    setViolations([])

    // Simulate AI generation
    setTimeout(() => {
      const newShifts: ScheduleShift[] = []
      const newLeaves: ScheduleLeave[] = []
      let idCounter = 1
      let leaveIdCounter = 1

      MOCK_EMPLOYEES.forEach((emp) => {
        const offDays = new Set<number>()
        if (leaveMode === 'random') {
          while (offDays.size < 2) {
            offDays.add(Math.floor(Math.random() * 7))
          }
        } else {
          offDays.add(0) // Sunday
        }

        weekDays.forEach((d, idx) => {
          if (offDays.has(idx % 7)) {
            newLeaves.push({
              id: leaveIdCounter++,
              employee_id: emp.id,
              employee_name: emp.name,
              date: d.date,
              is_override: false,
            })
            return
          }
          const shiftId = (idx % 3) + 1
          const shiftNames = ['', 'Pagi', 'Siang', 'Malam']
          newShifts.push({
            id: idCounter++,
            employee_id: emp.id,
            employee_name: emp.name,
            employee_role: emp.role_name,
            shift_template_id: shiftId,
            shift_name: shiftNames[shiftId] || 'Pagi',
            date: d.date,
            is_override: false,
          })
        })
      })

      // Simulate a rest-time violation for demo
      newShifts.push({
        id: idCounter++,
        employee_id: 5,
        employee_name: 'Rudi',
        employee_role: 'Supervisor',
        shift_template_id: 3,
        shift_name: 'Malam',
        date: weekDays[2]?.date || '2026-07-01',
        is_override: false,
      })

      setShifts(newShifts)
      setLeaves(newLeaves)

      // Check for violations
      const detected: ValidationViolation[] = []
      const shiftMap = new Map<string, { shift: ScheduleShift }>()
      newShifts.forEach((s) => {
        const key = `${s.employee_id}-${s.date}`
        shiftMap.set(key, { shift: s })
      })

      // Check consecutive day rest (simplified)
      newShifts.forEach((s) => {
        const prevDate = new Date(s.date)
        prevDate.setDate(prevDate.getDate() - 1)
        const prevKey = `${s.employee_id}-${prevDate.toISOString().split('T')[0]}`
        const prev = shiftMap.get(prevKey)
        if (prev && prev.shift.shift_template_id === 3 && s.shift_template_id === 1) {
          detected.push({
            type: 'rest_time_violation',
            severity: 'error',
            message: `${s.employee_name}: shift Malam ${prev.shift.date} → ${s.shift_name} ${s.date}, jeda kurang dari 12 jam`,
            employee_id: s.employee_id,
            employee_name: s.employee_name,
            date: s.date,
          })
        }
      })

      setViolations(detected)
      setIsGenerating(false)
    }, 1500)
  }

  const handleDrop = (targetEmployeeId: number, targetDate: string, item: any) => {
    setShifts((prev) =>
      prev.map((s) =>
        s.employee_id === item.employee_id && s.date === item.date
          ? { ...s, employee_id: targetEmployeeId, date: targetDate, is_override: true }
          : s,
      ),
    )
    setLeaves((prev) =>
      prev.map((l) =>
        l.employee_id === item.employee_id && l.date === item.date
          ? { ...l, employee_id: targetEmployeeId, date: targetDate, is_override: true }
          : l,
      ),
    )
  }

  const handlePublish = () => {
    setScheduleStatus('published')
  }

  return (
    <div className="flex h-screen">
      <Sidebar
        month={month}
        year={year}
        onPrevMonth={() => {
          if (month === 1) { setMonth(12); setYear(year - 1) }
          else setMonth(month - 1)
        }}
        onNextMonth={() => {
          if (month === 12) { setMonth(1); setYear(year + 1) }
          else setMonth(month + 1)
        }}
        leaveMode={leaveMode}
        onLeaveModeChange={setLeaveMode}
        isGenerating={isGenerating}
        onGenerate={handleGenerate}
        onExport={() => {}}
        onShare={() => {}}
        scheduleStatus={scheduleStatus ?? undefined}
      />

      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Top bar */}
        <header className="h-14 bg-card border-b border-border flex items-center justify-between px-6 shrink-0">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">
              Schedule {' '}
              {['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
                'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'][month - 1]}
              {' '}{year}
            </span>
            {scheduleStatus && (
              <Badge variant={scheduleStatus === 'published' ? 'secondary' : 'outline'}>
                {scheduleStatus === 'published' ? 'Published' : 'Draft'}
              </Badge>
            )}
            {scheduleStatus === 'draft' && (
              <span className="text-xs text-muted-foreground">Pending Review</span>
            )}
          </div>

          <div className="flex items-center gap-3">
            {scheduleStatus === 'draft' && (
              <Button variant="default" className="bg-emerald-600 hover:bg-emerald-700 gap-1.5" onClick={handlePublish}>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                Publish
              </Button>
            )}
            <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center text-sm font-medium">
              A
            </div>
          </div>
        </header>

        {/* Legend bar */}
        <div className="flex items-center justify-between px-6 py-3 bg-card border-b border-border">
          <div className="flex items-center gap-4">
            <h2 className="text-lg font-bold">Kalender Shift</h2>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><span className="w-2.5 h-2.5 rounded bg-blue-500" />Pagi</span>
              <span className="flex items-center gap-1"><span className="w-2.5 h-2.5 rounded bg-amber-500" />Siang</span>
              <span className="flex items-center gap-1"><span className="w-2.5 h-2.5 rounded bg-indigo-500" />Malam</span>
              <span className="flex items-center gap-1"><span className="w-2.5 h-2.5 rounded bg-muted-foreground/30" />Libur</span>
            </div>
          </div>
          {shifts.length > 0 && (
            <span className="text-xs text-muted-foreground">
              🎯 Fairness: SD 4.2 jam
            </span>
          )}
        </div>

        {/* Calendar */}
        <div className="flex-1 overflow-auto px-6 pb-6 pt-4">
          {shifts.length === 0 && !isGenerating ? (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
              <svg className="w-16 h-16 mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1}
                  d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <p className="text-lg font-medium mb-1">Belum ada jadwal</p>
              <p className="text-sm">Atur konfigurasi di sidebar lalu tekan <strong>Generate Jadwal</strong></p>
            </div>
          ) : isGenerating ? (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
              <span className="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin mb-4" />
              <p className="text-lg font-medium">AI sedang meng-generate jadwal...</p>
              <p className="text-sm">Memproses batch 1/1</p>
            </div>
          ) : (
            <>
              <CalendarGrid
                cells={cells}
                shifts={shifts}
                leaves={leaves}
                employees={MOCK_EMPLOYEES}
                weekDays={weekDays}
                onDrop={handleDrop}
              />

              {/* Violation banner */}
              {violations.length > 0 && (
                <div className="mt-4 p-4 rounded-lg bg-destructive/10 border border-destructive/30">
                  <div className="flex items-start gap-3">
                    <svg className="w-5 h-5 text-destructive mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                        d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <div>
                      <h4 className="text-sm font-semibold text-destructive">
                        {violations.length} Konflik Terdeteksi — Butuh Manual Fix
                      </h4>
                      {violations.map((v, i) => (
                        <p key={i} className="text-sm text-destructive/80 mt-0.5">
                          <strong>{v.employee_name}:</strong> {v.message}
                        </p>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Hint */}
              <div className="mt-4 text-xs text-muted-foreground flex items-center gap-4">
                <span>💡 Seret blok shift ke hari/karyawan lain untuk override manual</span>
                <span>|</span>
                <span>
                  <span className="inline-block w-3 h-3 border border-dashed border-muted-foreground/30 rounded align-middle mr-1" />
                  Belum di-generate
                </span>
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  )
}
