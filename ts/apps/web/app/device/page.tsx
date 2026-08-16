import { Suspense } from "react"
import { DeviceConfirm } from "@/components/device-confirm"

export const metadata = { title: "Approve device login" }

export default function DevicePage() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-8 p-6">
      <Suspense>
        <DeviceConfirm />
      </Suspense>
    </div>
  )
}
