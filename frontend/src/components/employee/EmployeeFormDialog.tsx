import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { Employee, CreateEmployeeRequest, UpdateEmployeeRequest } from '@/types'

interface EmployeeFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  employee?: Employee | null
  roles: { id: number; name: string }[]
  onSubmit: (data: CreateEmployeeRequest | UpdateEmployeeRequest) => Promise<void>
}

export function EmployeeFormDialog({
  open,
  onOpenChange,
  employee,
  roles,
  onSubmit,
}: EmployeeFormDialogProps) {
  const [name, setName] = useState('')
  const [roleId, setRoleId] = useState<string>('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState('')

  const isEdit = !!employee

  useEffect(() => {
    if (employee) {
      setName(employee.name)
      setRoleId(String(employee.role_id))
      setEmail(employee.email || '')
      setPhone(employee.phone || '')
    } else {
      setName('')
      setRoleId('')
      setEmail('')
      setPhone('')
    }
    setError('')
  }, [employee, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('Nama wajib diisi')
      return
    }
    if (!roleId) {
      setError('Role wajib dipilih')
      return
    }

    setIsSubmitting(true)
    try {
      if (isEdit) {
        await onSubmit({
          name: name.trim(),
          role_id: Number(roleId),
          email: email.trim() || undefined,
          phone: phone.trim() || undefined,
        } as UpdateEmployeeRequest)
      } else {
        await onSubmit({
          name: name.trim(),
          role_id: Number(roleId),
          email: email.trim() || undefined,
          phone: phone.trim() || undefined,
        } as CreateEmployeeRequest)
      }
      onOpenChange(false)
    } catch (err: any) {
      setError(err.message || 'Terjadi kesalahan')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Karyawan' : 'Tambah Karyawan'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Nama *</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Nama karyawan"
              autoFocus
            />
          </div>

          <div className="space-y-2">
            <Label>Role *</Label>
            <Select value={roleId} onValueChange={setRoleId}>
              <SelectTrigger>
                <SelectValue placeholder="Pilih role" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={String(role.id)}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="email@example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone">Telepon</Label>
            <Input
              id="phone"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="08xxxxxxxxxx"
            />
          </div>

          {error && (
            <p className="text-sm text-destructive">{error}</p>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Batal
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Menyimpan...' : isEdit ? 'Simpan' : 'Tambah'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
