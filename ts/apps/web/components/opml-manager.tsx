"use client";

import { useState, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { FileDown, Upload, FileText, Loader2, CheckCircle2 } from "lucide-react";
import { Button } from "@workspace/ui/components/button";
import { getClient, exportOpml, importOpml } from "@/lib/planetary";
import { getApiErrorMessage } from "@/lib/errors";

interface ImportResult {
  imported: number;
  skipped: number;
  failed: number;
  feed_ids: number[];
  errors?: string[];
}

/** OPML import/export manager. Export downloads all subscriptions as OPML; import subscribes from an OPML file. */
export function OpmlManager() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [exporting, setExporting] = useState(false);
  const [importing, setImporting] = useState(false);
  const [result, setResult] = useState<ImportResult | null>(null);

  async function handleExport() {
    setExporting(true);
    try {
      const { data, error } = await exportOpml({
        client: await getClient(),
      });
      if (error) throw error;
      const xml = data as unknown as string;
      const blob = new Blob([xml], { type: "application/xml" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "planetary-subscriptions.opml";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success("Exported subscriptions as OPML");
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not export OPML"));
    } finally {
      setExporting(false);
    }
  }

  async function handleImport(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setImporting(true);
    setResult(null);
    try {
      const { data, error } = await importOpml({
        client: await getClient(),
        body: file,
      });
      if (error) throw error;
      const res = data as unknown as ImportResult;
      setResult(res);
      await queryClient.invalidateQueries({ queryKey: ["entries"] });
      await queryClient.invalidateQueries({ queryKey: ["feeds"] });
      await queryClient.invalidateQueries({ queryKey: ["categories"] });
      toast.success(
        `Imported ${res.imported} feed${res.imported === 1 ? "" : "s"}, skipped ${res.skipped}`,
      );
    } catch (err) {
      toast.error(getApiErrorMessage(err, "Could not import OPML"));
    } finally {
      setImporting(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:px-6">
      <header className="flex items-center gap-2">
        <FileText className="size-5 text-muted-foreground" />
        <h1 className="font-serif text-2xl font-bold tracking-tight">OPML</h1>
      </header>
      <p className="mt-2 text-sm text-muted-foreground">
        Import or export your feed subscriptions as an OPML file. OPML is the standard format for
        sharing RSS subscriptions between readers.
      </p>

      <div className="mt-6 space-y-4">
        <div className="flex items-center justify-between rounded-lg border border-border p-4">
          <div>
            <h3 className="text-sm font-medium">Export subscriptions</h3>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Download all your feeds and categories as an OPML file.
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={handleExport} disabled={exporting}>
            {exporting ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <FileDown className="size-3.5" />
            )}
            Export
          </Button>
        </div>

        <div className="flex items-center justify-between rounded-lg border border-border p-4">
          <div>
            <h3 className="text-sm font-medium">Import subscriptions</h3>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Upload an OPML file to subscribe to all feeds. Categories are created automatically.
              Existing subscriptions are skipped.
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => fileInputRef.current?.click()}
            disabled={importing}
          >
            {importing ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <Upload className="size-3.5" />
            )}
            Import
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".opml,.xml,application/xml,text/xml"
            onChange={handleImport}
            className="hidden"
          />
        </div>

        {result && (
          <div className="rounded-lg border border-border bg-muted/30 p-4">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="size-4 text-green-600" />
              <h3 className="text-sm font-medium">Import complete</h3>
            </div>
            <dl className="mt-3 grid grid-cols-3 gap-4 text-center">
              <div>
                <dt className="text-xs text-muted-foreground">Imported</dt>
                <dd className="mt-1 text-lg font-semibold">{result.imported}</dd>
              </div>
              <div>
                <dt className="text-xs text-muted-foreground">Skipped</dt>
                <dd className="mt-1 text-lg font-semibold">{result.skipped}</dd>
              </div>
              <div>
                <dt className="text-xs text-muted-foreground">Failed</dt>
                <dd className="mt-1 text-lg font-semibold">{result.failed}</dd>
              </div>
            </dl>
            {result.errors && result.errors.length > 0 && (
              <details className="mt-3">
                <summary className="cursor-pointer text-xs text-muted-foreground hover:text-foreground">
                  Show {result.errors.length} error{result.errors.length === 1 ? "" : "s"}
                </summary>
                <ul className="mt-2 space-y-1">
                  {result.errors.map((err, i) => (
                    <li key={i} className="text-xs text-destructive">
                      {err}
                    </li>
                  ))}
                </ul>
              </details>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
