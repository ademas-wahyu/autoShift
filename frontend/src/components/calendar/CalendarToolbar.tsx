import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface CalendarToolbarProps {
  viewMode: 'week' | 'month'
  onViewModeChange: (mode: 'week' | 'month') => void
  selectedWeek: number
  totalWeeks: number
  onPrevWeek: () => void
  onNextWeek: () => void
  weekRange: string
  roles: string[]
  selectedRole: string
  onRoleChange: (role: string) => void
  employeeCount: number
}

export function CalendarToolbar({
  viewMode,
  onViewModeChange,
  selectedWeek,
  totalWeeks,
  onPrevWeek,
  onNextWeek,
  weekRange,
  roles,
  selectedRole,
  onRoleChange,
  employeeCount,
}: CalendarToolbarProps) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      {/* Left: View mode toggle */}
      <div className="flex items-center gap-2">
        <div className="inline-flex rounded-lg border bg-muted/50 p-0.5">
          <button
            onClick={() => onViewModeChange('week')}
            className={cn(
              'px-3 py-1.5 text-sm font-medium rounded-md transition-all',
              viewMode === 'week'
                ? 'bg-background shadow-sm text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            Mingguan
          </button>
          <button
            onClick={() => onViewModeChange('month')}
            className={cn(
              'px-3 py-1.5 text-sm font-medium rounded-md transition-all',
              viewMode === 'month'
                ? 'bg-background shadow-sm text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            Bulanan
          </button>
        </div>

        {/* Week navigation (only in week view) */}
        {viewMode === 'week' && (
          <div className="flex items-center gap-1 ml-2">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={onPrevWeek}
              disabled={selectedWeek === 0}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </Button>
            <div className="px-3 py-1 rounded-md bg-muted/50 text-sm font-medium min-w-[140px] text-center">
              {weekRange}
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={onNextWeek}
              disabled={selectedWeek === totalWeeks - 1}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </Button>
          </div>
        )}
      </div>

      {/* Right: Role filter tabs */}
      <div className="flex items-center gap-2">
        <div className="flex items-center gap-1 overflow-x-auto pb-1">
          {roles.map((role) => (
            <button
              key={role}
              onClick={() => onRoleChange(role)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium rounded-md transition-all whitespace-nowrap',
                selectedRole === role
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              {role}
              {role !== 'Semua' && (
                <span className="ml-1.5 text-xs opacity-70">
                  ({employeeCount})
                </span>
              )}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
