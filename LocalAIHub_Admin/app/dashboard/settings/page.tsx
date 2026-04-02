'use client'

import { useState } from 'react'
import { LuDownload, LuUpload, LuCheck, LuX, LuFileJson } from 'react-icons/lu'
import { FiSettings } from 'react-icons/fi'
import { api } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
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
      const params = new URLSearchParams()
      Object.entries(exportModules).forEach(([key, value]) => {
        params.set(key, String(value))
      })
      const data = await api.exportConfig(params.toString())
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

  function clearFile() {
    setImportFile(null)
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
              <div className="space-y-3">
                <Switch
                  checked={exportModules.providers}
                  onChange={(e) => setExportModules({ ...exportModules, providers: e.target.checked })}
                  label="上游"
                />
                <Switch
                  checked={exportModules.virtual_models}
                  onChange={(e) => setExportModules({ ...exportModules, virtual_models: e.target.checked })}
                  label="模型"
                />
                <Switch
                  checked={exportModules.bindings}
                  onChange={(e) => setExportModules({ ...exportModules, bindings: e.target.checked })}
                  label="绑定"
                />
                <Switch
                  checked={exportModules.api_clients}
                  onChange={(e) => setExportModules({ ...exportModules, api_clients: e.target.checked })}
                  label="API Key"
                />
                <Switch
                  checked={exportModules.provider_keys}
                  onChange={(e) => setExportModules({ ...exportModules, provider_keys: e.target.checked })}
                  label="Provider Keys"
                />
              </div>
              <Button onClick={handleExport} loading={exporting}>
                <LuDownload className="h-4 w-4 mr-1" />
                导出配置
              </Button>
            </div>

            <div className="space-y-4">
              <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>导入配置</h3>
              <div className="space-y-4">
                <div>
                  <label className="text-sm mb-2 block" style={{ color: 'var(--foreground)' }}>配置文件</label>
                  {!importFile ? (
                    <label className="flex items-center justify-center h-12 rounded-[10px] border-2 border-dashed cursor-pointer transition-colors hover:bg-[var(--accent)]" style={{ borderColor: 'var(--border)' }}>
                      <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--muted-foreground)' }}>
                        <LuFileJson className="h-4 w-4" />
                        <span>点击选择 JSON 文件</span>
                      </div>
                      <input type="file" accept=".json" onChange={handleFileChange} className="hidden" />
                    </label>
                  ) : (
                    <div className="flex items-center justify-between h-12 rounded-[10px] border px-4" style={{ borderColor: 'var(--border)', background: 'var(--input)' }}>
                      <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                        <LuFileJson className="h-4 w-4" />
                        <span className="truncate max-w-[200px]">{importFile.name}</span>
                      </div>
                      <button onClick={clearFile} className="text-sm hover:opacity-70" style={{ color: 'var(--muted-foreground)' }}>清除</button>
                    </div>
                  )}
                </div>
                <Switch
                  checked={importOptions.overwrite_existing}
                  onChange={(e) => setImportOptions({ ...importOptions, overwrite_existing: e.target.checked })}
                  label="覆盖已存在"
                />
                <Switch
                  checked={importOptions.skip_invalid}
                  onChange={(e) => setImportOptions({ ...importOptions, skip_invalid: e.target.checked })}
                  label="跳过无效项"
                />
                <Switch
                  checked={importOptions.dry_run}
                  onChange={(e) => setImportOptions({ ...importOptions, dry_run: e.target.checked })}
                  label="试运行（不实际导入）"
                />
              </div>
              <Button onClick={handleImport} loading={importing} disabled={!importFile}>
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
