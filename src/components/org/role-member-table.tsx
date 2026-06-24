import { useState } from 'react'
import type { Member, Role } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface RoleMemberTableProps {
  role: Role
  members: Member[]
  onRemoveMember: (member: Member) => void
  onAddMember: () => void
}

export function RoleMemberTable({
  role,
  members,
  onRemoveMember,
  onAddMember,
}: RoleMemberTableProps) {
  const [search, setSearch] = useState('')

  const filtered = members.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="flex-1 flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-foreground">{role.name}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {role.type === 'preset' ? '系统预设角色' : '自定义角色'} · {members.length} 名成员
          </p>
        </div>
        <Button onClick={onAddMember}>添加角色成员</Button>
      </div>

      <div className="mb-3">
        <Input
          type="text"
          placeholder="搜索成员..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-xs"
        />
      </div>

      <div className="border border-border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="px-4">姓名</TableHead>
              <TableHead className="px-4">角色</TableHead>
              <TableHead className="px-4 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} className="px-4 py-6 text-center text-muted-foreground">
                  暂无成员
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((member) => (
                <TableRow key={member.id}>
                  <TableCell className="px-4">{member.name}</TableCell>
                  <TableCell className="px-4">
                    <div className="flex flex-wrap gap-1">
                      {member.roles.map((r) => (
                        <Badge key={r} variant="secondary">
                          {r}
                        </Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="px-4 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      onClick={() => onRemoveMember(member)}
                    >
                      移除
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
