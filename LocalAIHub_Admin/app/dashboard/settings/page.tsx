'use client'

import { useState } from 'react'
import { LuDownload, LuUpload, LuFileJson, LuCheck, LuX, LuLoader } from 'react-icons/lu'
import { FiSettings } from 'react-icons/fi'
import { api } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
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
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>导出配置</h3>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={exportModules.providers}
                    onChange={(e) => setExportModules({ ...exportModules, providers: e.target.checked })}
                    className="rounded"
                  />
                  上游服务商
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={exportModules.virtual_models}
                    onChange={(e) => setExportModules({ ...exportModules, virtual_models: e.target.checked })}
                    className="rounded"
                  />
                  虚拟模型
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={exportModules.bindings}
                    onChange={(e) => setExportModules({ ...exportModules, bindings: e.target.checked })}
                    className="rounded"
                  />
                  模型绑定
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={exportModules.api_clients}
                    onChange={(e) => setExportModules({ ...exportModules, api_clients: e.target.checked })}
                    className="rounded"
                  />
                  API Key
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={exportModules.provider_keys}
                    onChange={(e) => setExportModules({ ...exportModules, provider_keys: e.target.checked })}
                    className="rounded"
                  />
                  上游 Key（仅掩码）
                </label>
              </div>
              <Button onClick={handleExport} loading={exporting}>
                <LuDownload className="h-4 w-4 mr-1" />
                导出配置
              </Button>
            </div>

            <div className="space-y-4">
              <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>导入配置</h3>
              <div>
                <label className="text-sm mb-1 block" style={{ color: 'var(--foreground)' }}>选择文件</label>
                <div
                  className="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer hover:border-[var(--foreground)] transition-colors"
                  style={{ borderColor: 'var(--border)' }}
                  onClick={() => document.getElementById('import-file')?.click()}
                >
                  <input
                    id="import-file"
                    type="file"
                    accept=".json"
                    className="hidden"
                    onChange={handleFileChange}
                  />
                  <LuFileJson className="h-8 w-8 mx-auto mb-2" style={{ color: 'var(--muted-foreground)' }} />
                  <div className="text-sm" style={{ color: 'var(--muted-foreground)' }}>
                    {importFile ? importFile.name : '点击选择 JSON 文件'}
                  </div>
                </div>
              </div>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={importOptions.overwrite_existing}
                    onChange={(e) => setImportOptions({ ...importOptions, overwrite_existing: e.target.checked })}
                    className="rounded"
                  />
                  覆盖已存在配置
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={importOptions.skip_invalid}
                    onChange={(e) => setImportOptions({ ...importOptions, skip_invalid: e.target.checked })}
                    className="rounded"
                  />
                  跳过无效记录
                </label>
                <label className="flex items-center gap-2 text-sm" style={{ color: 'var(--foreground)' }}>
                  <input
                    type="checkbox"
                    checked={importOptions.dry_run}
                    onChange={(e) => setImportOptions({ ...importOptions, dry_run: e.target.checked })}
                    className="rounded"
                  />
                  试运行（不实际导入）
                </label>
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
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                {importResult.summary && Object.entries(importResult.summary).map(([key, value]) => {
                  if (typeof value === 'number' && value > 0) {
                    return (
                      <div key={key} className="flex items-center gap-2 text-sm">
                        <LuCheck className="h-4 w-4 text-emerald-500" />
                        <span style={{ color: 'var(--muted-foreground)' }}>{key}: </span>
                        <span style={{ color: 'var(--foreground)' }}>{value}</span>
                      </div>
                    )
                  }
                  return null
                })}
              </div>
              {importResult.errors?.length > 0 && (
                <div className="mt-3 space-y-1">
                  {importResult.errors.map((err: string, i: number) => (
                    <div key={i} className="flex items-center gap-2 text-sm text-red-500">
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
