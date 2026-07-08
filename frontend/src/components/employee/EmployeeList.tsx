import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { EmployeeFormDialog } from './EmployeeFormDialog'
import { api } from '@/lib/api'
import type { Employee, CreateEmployeeRequest, UpdateEmployeeRequest, EmployeeRole } from '@/types'

const ROLES: Pick<EmployeeRole, 'id' | 'name'>[] = [
  { id: 1, name: 'Supervisor' },
  { id: 2, name: 'Staff' },
  { id: 3, name: 'Intern' },
]

interface EmployeeListProps {
  onEmployeesChange?: (employees: Employee[]) => void
}

export function EmployeeList({ onEmployeesChange }: EmployeeListProps) {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [filterRole, setFilterRole] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null)

  const fetchEmployees = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      const params: { is_active?: boolean; role_id?: number } = {}
      if (filterStatus === 'active') params.is_active = true
      if (filterStatus === 'inactive') params.is_active = false
      if (filterRole !== 'all') params.role_id = Number(filterRole)

      const res = await api.listEmployees(params)
      setEmployees(res.data)
      onEmployeesChange?.(res.data)
    } catch (err: any) {
      setError(err.message || 'Gagal memuat data karyawan')
    } finally {
      setIsLoading(false)
    }
  }, [filterRole, filterStatus, onEmployeesChange])

  useEffect(() => {
    fetchEmployees()
  }, [fetchEmployees])

  const handleCreate = async (data: CreateEmployeeRequest | UpdateEmployeeRequest) => {
    await api.createEmployee(data as CreateEmployeeRequest)
    await fetchEmployees()
  }

  const handleUpdate = async (data: CreateEmployeeRequest | UpdateEmployeeRequest) => {
    if (!editingEmployee) return
    await api.updateEmployee(editingEmployee.id, data as UpdateEmployeeRequest)
    setEditingEmployee(null)
    await fetchEmployees()
  }

  const handleDelete = async (employee: Employee) => {
    if (!confirm(`Nonaktifkan karyawan "${employee.name}"?`)) return
    try {
      await api.deleteEmployee(employee.id)
      await fetchEmployees()
    } catch (err: any) {
      alert(err.message || 'Gagal menghapus karyawan')
    }
  }

  const handleEdit = (employee: Employee) => {
    setEditingEmployee(employee)
    setDialogOpen(true)
  }

  const handleAdd = () => {
    setEditingEmployee(null)
    setDialogOpen(true)
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-bold">Daftar Karyawan</h2>
        <Button onClick={handleAdd} className="gap-2">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Tambah
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <Select value={filterRole} onValueChange={setFilterRole}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Semua Role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Role</SelectItem>
            {ROLES.map((role) => (
              <SelectItem key={role.id} value={String(role.id)}>
                {role.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={filterStatus} onValueChange={setFilterStatus}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Semua Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Status</SelectItem>
            <SelectItem value="active">Aktif</SelectItem>
            <SelectItem value="inactive">Nonaktif</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Error */}
      {error && (
        <div className="p-3 rounded-md bg-destructive/10 text-destructive text-sm">
          {error}
        </div>
      )}

      {/* Table */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <span className="w-6 h-6 border-2 border-primary/30 border-t-primary rounded-full animate-spin mr-3" />
          Memuat data...
        </div>
      ) : employees.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <svg className="w-12 h-12 mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1}
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <p className="font-medium">Belum ada karyawan</p>
          <p className="text-sm mt-1">Klik tombol "Tambah" untuk menambahkan karyawan baru</p>
        </div>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Telepon</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-[100px]">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {employees.map((emp) => (
                <TableRow key={emp.id}>
                  <TableCell className="font-medium">{emp.name}</TableCell>
                  <TableCell>{emp.role?.name || '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{emp.email || '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{emp.phone || '-'}</TableCell>
                  <TableCell>
                    <Badge variant={emp.is_active ? 'secondary' : 'outline'}>
                      {emp.is_active ? 'Aktif' : 'Nonaktif'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => handleEdit(emp)}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={() => handleDelete(emp)}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Count */}
      {!isLoading && employees.length > 0 && (
        <p className="text-xs text-muted-foreground text-right">
          {employees.length} karyawan
        </p>
      )}

      {/* Form Dialog */}
      <EmployeeFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        employee={editingEmployee}
        roles={ROLES}
        onSubmit={editingEmployee ? handleUpdate : handleCreate}
      />
    </div>
  )
}
