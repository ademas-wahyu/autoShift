import { cn } from '@/lib/utils'
import type { CalendarCell, ScheduleShift, ScheduleLeave } from '@/types'

interface CalendarGridProps {
  cells: CalendarCell[]
  shifts: ScheduleShift[]
  leaves: ScheduleLeave[]
  employees: { id: number; name: string; role_name: string }[]
  weekDays: { label: string; date: string }[]
  viewMode: 'week' | 'month'
  selectedWeek: number
  onDrop?: (employeeId: number, date: string, item: any) => void
}

export function CalendarGrid({
  cells,
  shifts,
  leaves,
  employees,
  weekDays,
  viewMode,
  selectedWeek,
  onDrop,
}: CalendarGridProps) {
  // Get the days to display based on view mode
  const displayDays = viewMode === 'week'
    ? weekDays.slice(selectedWeek * 7, (selectedWeek + 1) * 7)
    : weekDays

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
    ;(e.currentTarget as HTMLElement).classList.add('ring-2', 'ring-primary/50')
  }

  const handleDragLeave = (e: React.DragEvent) => {
    ;(e.currentTarget as HTMLElement).classList.remove('ring-2', 'ring-primary/50')
  }

  const handleDrop = (e: React.DragEvent, employeeId: number, date: string) => {
    e.preventDefault()
    ;(e.currentTarget as HTMLElement).classList.remove('ring-2', 'ring-primary/50')
    try {
      const item = JSON.parse(e.dataTransfer.getData('application/json'))
      onDrop?.(employeeId, date, item)
    } catch { /* ignore */ }
  }

  const isWeekView = viewMode === 'week'
  const gridCols = isWeekView
    ? 'grid-cols-[160px_repeat(7,1fr)]'
    : 'grid-cols-[140px_repeat(7,1fr)]'

  return (
    <div className="overflow-auto rounded-xl border bg-card shadow-sm">
      {/* Header row */}
      <div className={cn('grid sticky top-0 z-10 bg-muted/80 backdrop-blur-sm border-b', gridCols)}>
        <div className={cn(
          'p-3 text-xs font-semibold text-muted-foreground flex items-center',
          isWeekView ? 'pl-4' : 'pl-3'
        )}>
          Karyawan
        </div>
        {displayDays.map((d) => {
          const c = cell(d.date)
          const isToday = c?.isToday
          const isHoliday = c?.isHoliday
          return (
            <div
              key={d.date}
              className={cn(
                'p-2 text-center border-l transition-colors',
                isHoliday && 'bg-red-50/80',
                isToday && 'bg-primary/5',
                isWeekView && 'py-3',
              )}
            >
              <div className={cn(
                'text-xs font-semibold uppercase tracking-wider',
                isToday ? 'text-primary' : 'text-muted-foreground'
              )}>
                {d.label}
              </div>
              <div className={cn(
                'font-bold',
                isWeekView ? 'text-2xl mt-1' : 'text-lg',
                isHoliday ? 'text-red-500' : isToday ? 'text-primary' : 'text-foreground',
              )}>
                {new Date(d.date).getDate()}
              </div>
              {isHoliday && (
                <div className="text-[10px] text-red-400 font-medium mt-0.5 truncate" title={c.holidayName}>
                  {c.holidayName || 'Libur'}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Employee rows */}
      <div className="divide-y">
        {employees.map((emp) => (
          <div key={emp.id} className={cn('grid', gridCols)}>
            {/* Employee info */}
            <div className={cn(
              'flex items-center gap-2.5 border-r bg-card sticky left-0 z-5',
              isWeekView ? 'p-3 pl-4' : 'p-2 pl-3'
            )}>
              <div className={cn(
                'rounded-full bg-primary/10 flex items-center justify-center font-semibold text-primary shrink-0',
                isWeekView ? 'w-9 h-9 text-sm' : 'w-7 h-7 text-xs'
              )}>
                {emp.name[0]}
              </div>
              <div className="min-w-0">
                <div className={cn(
                  'font-medium truncate',
                  isWeekView ? 'text-sm' : 'text-xs'
                )}>
                  {emp.name}
                </div>
                <div className="text-[11px] text-muted-foreground">{emp.role_name}</div>
              </div>
            </div>

            {/* Day cells */}
            {displayDays.map((d) => {
              const s = getShiftFor(emp.id, d.date)
              const l = getLeaveFor(emp.id, d.date)
              const c = cell(d.date)
              const isEmpty = !s && !l

              return (
                <div
                  key={`${emp.id}-${d.date}`}
                  className={cn(
                    'border-l transition-colors relative group',
                    c?.isHoliday && 'bg-red-50/30',
                    c?.isToday && 'bg-primary/[0.02]',
                    isEmpty && 'bg-muted/20',
                    isWeekView ? 'min-h-[80px] p-2' : 'min-h-[64px] p-1.5',
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
                        'h-full rounded-lg transition-all cursor-grab active:cursor-grabbing',
                        'hover:shadow-md hover:scale-[1.02] active:scale-[0.98]',
                        'border-l-4',
                        s.shift_template_id === 1 && 'bg-blue-50 border-l-blue-500 hover:bg-blue-100/80',
                        s.shift_template_id === 2 && 'bg-amber-50 border-l-amber-500 hover:bg-amber-100/80',
                        s.shift_template_id === 3 && 'bg-indigo-50 border-l-indigo-500 hover:bg-indigo-100/80',
                        isWeekView ? 'px-3 py-2' : 'px-2 py-1.5',
                      )}
                    >
                      <div className="flex flex-col h-full justify-center">
                        <span className={cn(
                          'font-semibold truncate',
                          isWeekView ? 'text-sm' : 'text-xs'
                        )}>
                          {s.shift_name}
                        </span>
                        {isWeekView && (
                          <span className="text-[11px] text-muted-foreground mt-0.5">
                            {s.shift_template_id === 1 && '08:00 - 16:00'}
                            {s.shift_template_id === 2 && '16:00 - 00:00'}
                            {s.shift_template_id === 3 && '22:00 - 06:00'}
                          </span>
                        )}
                        {s.is_override && (
                          <span className="text-[10px] text-muted-foreground italic mt-0.5">
                            Manual
                          </span>
                        )}
                      </div>
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
                      className={cn(
                        'h-full rounded-lg bg-muted/80 border-l-4 border-l-muted-foreground/20',
                        'cursor-grab active:cursor-grabbing hover:shadow-md transition-all',
                        'flex items-center justify-center',
                        isWeekView ? 'px-3 py-2' : 'px-2 py-1.5',
                      )}
                    >
                      <span className={cn(
                        'text-muted-foreground font-medium',
                        isWeekView ? 'text-sm' : 'text-xs'
                      )}>
                        Libur
                      </span>
                    </div>
                  )}
                  {isEmpty && (
                    <div className="h-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                      <span className="text-xs text-muted-foreground/40">—</span>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
