import { cn } from '@/lib/utils'
import type { CalendarCell, ScheduleShift, ScheduleLeave } from '@/types'

interface CalendarGridProps {
  cells: CalendarCell[]
  shifts: ScheduleShift[]
  leaves: ScheduleLeave[]
  employees: { id: number; name: string; role_name: string }[]
  weekDays: { label: string; date: string }[]
  onDrop?: (employeeId: number, date: string, item: any) => void
}

export function CalendarGrid({ cells, shifts, leaves, employees, weekDays, onDrop }: CalendarGridProps) {
  const getShiftFor = (empId: number, date: string) =>
    shifts.find((s) => s.employee_id === empId && s.date === date)

  const getLeaveFor = (empId: number, date: string) =>
    leaves.find((l) => l.employee_id === empId && l.date === date)

  const cell = (date: string) =>
    cells.find((c) => c.date === date)

  const handleDragStart = (e: React.DragEvent, item: any) => {
    e.dataTransfer.setData('application/json', JSON.stringify(item))
    ;(e.currentTarget as HTMLElement).classList.add('opacity-50')
  }

  const handleDragEnd = (e: React.DragEvent) => {
    ;(e.currentTarget as HTMLElement).classList.remove('opacity-50')
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    ;(e.currentTarget as HTMLElement).classList.add('ring-2', 'ring-blue-400')
  }

  const handleDragLeave = (e: React.DragEvent) => {
    ;(e.currentTarget as HTMLElement).classList.remove('ring-2', 'ring-blue-400')
  }

  const handleDrop = (e: React.DragEvent, employeeId: number, date: string) => {
    e.preventDefault()
    ;(e.currentTarget as HTMLElement).classList.remove('ring-2', 'ring-blue-400')
    try {
      const item = JSON.parse(e.dataTransfer.getData('application/json'))
      onDrop?.(employeeId, date, item)
    } catch { /* ignore */ }
  }

  return (
    <div className="overflow-auto rounded-lg border">
      {/* Header row */}
      <div className="grid grid-cols-[180px_repeat(7,1fr)] bg-muted/50 border-b">
        <div className="p-3 text-xs font-semibold text-muted-foreground">Karyawan</div>
        {weekDays.map((d) => {
          const c = cell(d.date)
          return (
            <div
              key={d.date}
              className={cn(
                'p-2 text-center border-l',
                c?.isHoliday && 'bg-red-50',
                c?.isToday && 'bg-blue-50',
              )}
            >
              <div className="text-xs font-semibold text-muted-foreground">{d.label}</div>
              <div className={cn(
                'text-lg font-bold',
                c?.isHoliday ? 'text-red-500' : 'text-foreground',
                c?.isToday && 'text-blue-600',
              )}>
                {new Date(d.date).getDate()}
              </div>
              {c?.isHoliday && (
                <div className="text-[10px] text-red-400 font-medium truncate" title={c.holidayName}>
                  {c.holidayName || 'Libur'}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Employee rows */}
      {employees.map((emp) => (
        <div key={emp.id} className="grid grid-cols-[180px_repeat(7,1fr)] border-b last:border-b-0">
          {/* Employee info */}
          <div className="p-2 flex items-center gap-2 border-r bg-background">
            <div className="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-semibold text-primary shrink-0">
              {emp.name[0]}
            </div>
            <div className="min-w-0">
              <div className="text-sm font-medium truncate">{emp.name}</div>
              <div className="text-xs text-muted-foreground">{emp.role_name}</div>
            </div>
          </div>

          {/* Day cells */}
          {weekDays.map((d) => {
            const s = getShiftFor(emp.id, d.date)
            const l = getLeaveFor(emp.id, d.date)
            const c = cell(d.date)
            const isEmpty = !s && !l

            return (
              <div
                key={`${emp.id}-${d.date}`}
                className={cn(
                  'p-1 border-l min-h-[68px] transition-colors',
                  c?.isHoliday && 'bg-red-50/50',
                  isEmpty && 'bg-muted/20',
                )}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={(e) => handleDrop(e, emp.id, d.date)}
              >
                {s && (
                  <div
                    draggable
                    onDragStart={(e) => handleDragStart(e, {
                      type: 'shift',
                      schedule_shift_id: s.id,
                      employee_id: s.employee_id,
                      shift_template_id: s.shift_template_id,
                      date: s.date,
                    })}
                    onDragEnd={handleDragEnd}
                    className={cn(
                      'h-full rounded px-2 py-1.5 flex flex-col cursor-grab active:cursor-grabbing',
                      'hover:shadow-md transition-all border-l-4',
                      s.shift_template_id === 1 && 'bg-blue-50 border-blue-500',
                      s.shift_template_id === 2 && 'bg-amber-50 border-amber-500',
                      s.shift_template_id === 3 && 'bg-indigo-50 border-indigo-500',
                    )}
                  >
                    <span className="text-xs font-medium truncate">{s.shift_name}</span>
                    {s.is_override && (
                      <span className="text-[10px] text-muted-foreground italic">Manual</span>
                    )}
                  </div>
                )}
                {l && (
                  <div
                    draggable
                    onDragStart={(e) => handleDragStart(e, {
                      type: 'leave',
                      schedule_leave_id: l.id,
                      employee_id: l.employee_id,
                      date: l.date,
                    })}
                    onDragEnd={handleDragEnd}
                    className="h-full rounded px-2 py-1.5 flex items-center bg-muted border-l-4 border-muted-foreground/30 cursor-grab active:cursor-grabbing hover:shadow-md transition-all"
                  >
                    <span className="text-xs text-muted-foreground italic">Libur</span>
                  </div>
                )}
                {isEmpty && (
                  <div className="h-full flex items-center justify-center">
                    <span className="text-xs text-muted-foreground/30">—</span>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      ))}
    </div>
  )
}
