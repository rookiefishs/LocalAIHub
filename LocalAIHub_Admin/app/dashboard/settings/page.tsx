'use client'

import { useState } from 'react'
import { LuDownload, LuUpload, LuFileJson, LuCheck, LuX } from 'react-icons/lu'
import { FiSettings } from 'react-icons/fi'
import { api } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useToast } from '@/components/ui/toast'

export default function SettingsPage() {
  const [exportModules, setExportModules] = useState({
    providers: true,
    virtual_models: true,
    bindings: true,
    api_clients: true,
    provider_keys: true,
  })
  const [exporting, setExporting] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importOptions, setImportOptions] = useState({
    overwrite_existing: true,
    skip_invalid: true,
    dry_run: false,
  })
  const [importResult, setImportResult] = useState<any>(null)
  const { showSuccess, showError } = useToast()

  async function handleExport() {
    setExporting(true)
    try {
      const data = await api.exportConfig()
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `localaihub-config-${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
      showSuccess('配置导出成功')
    } catch (err) {
      showError(err instanceof Error ? err.message : '导出失败')
    } finally {
      setExporting(false)
    }
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) {
      setImportFile(file)
      setImportResult(null)
    }
  }

  async function handleImport() {
    if (!importFile) {
      showError('请选择配置文件')
      return
    }

    setImporting(true)
    try {
      const text = await importFile.text()
      const config = JSON.parse(text)
      const result = await api.importConfig({
        config,
        options: importOptions,
      })
      setImportResult(result)
      showSuccess('配置导入成功')
    } catch (err) {
      showError(err instanceof Error ? err.message : '导入失败')
    } finally {
      setImporting(false)
    }
  }

  return (
    <div className="space-y-4">
      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b px-6 py-4" style={{ borderColor: 'var(--border)' }}>
          <div className="flex items-center gap-2">
            <FiSettings className="h-4 w-4" style={{ color: 'var(--muted-foreground)' }} />
            <span className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>配置导入导出</span>
          </div>
        </div>
        <CardContent className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div className="space-y-4">
              <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>导出配置</h3>
              <div className="space-y-4">
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>上游</label>
                  <input
                    type="checkbox"
                    checked={exportModules.providers}
                    onChange={(e) => setExportModules({ ...exportModules, providers: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>模型</label>
                  <input
                    type="checkbox"
                    checked={exportModules.virtual_models}
                    onChange={(e) => setExportModules({ ...exportModules, virtual_models: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>绑定</label>
                  <input
                    type="checkbox"
                    checked={exportModules.bindings}
                    onChange={(e) => setExportModules({ ...exportModules, bindings: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>Key</label>
                  <input
                    type="checkbox"
                    checked={exportModules.api_clients}
                    onChange={(e) => setExportModules({ ...exportModules, api_clients: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
              </div>
              <Button onClick={handleExport} loading={exporting}>
                <LuDownload className="h-4 w-4 mr-1" />
                导出配置
              </Button>
            </div>

            <div className="space-y-4">
              <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>导入配置</h3>
              <div className="space-y-4">
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>文件</label>
                  <Input
                    type="file"
                    accept=".json"
                    onChange={handleFileChange}
                    className="flex-1"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>覆盖</label>
                  <input
                    type="checkbox"
                    checked={importOptions.overwrite_existing}
                    onChange={(e) => setImportOptions({ ...importOptions, overwrite_existing: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>跳过</label>
                  <input
                    type="checkbox"
                    checked={importOptions.skip_invalid}
                    onChange={(e) => setImportOptions({ ...importOptions, skip_invalid: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <label className="w-16 text-sm" style={{ color: 'var(--foreground)' }}>试运行</label>
                  <input
                    type="checkbox"
                    checked={importOptions.dry_run}
                    onChange={(e) => setImportOptions({ ...importOptions, dry_run: e.target.checked })}
                    className="h-4 w-4"
                  />
                </div>
              </div>
              <Button onClick={handleImport} loading={importing}>
                <LuUpload className="h-4 w-4 mr-1" />
                导入配置
              </Button>
            </div>
          </div>

          {importResult && (
            <div className="mt-6 p-4 rounded-lg border" style={{ borderColor: 'var(--border)', background: 'rgba(0,0,0,0.1)' }}>
              <h4 className="text-sm font-medium mb-3" style={{ color: 'var(--foreground)' }}>导入结果</h4>
              {importResult.summary && (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-3">
                  {Object.entries(importResult.summary).map(([key, value]) => {
                    if (typeof value === 'number' && value > 0) {
                      return (
                        <div key={key} className="flex items-center gap-2 text-sm">
                          <LuCheck className="h-4 w-4" style={{ color: 'var(--success)' }} />
                          <span style={{ color: 'var(--muted-foreground)' }}>{key}: </span>
                          <span style={{ color: 'var(--foreground)' }}>{value}</span>
                        </div>
                      )
                    }
                    return null
                  })}
                </div>
              )}
              {importResult.errors?.length > 0 && (
                <div className="space-y-1">
                  {importResult.errors.map((err: string, i: number) => (
                    <div key={i} className="flex items-center gap-2 text-sm" style={{ color: 'var(--danger)' }}>
                      <LuX className="h-4 w-4" />
                      {err}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
