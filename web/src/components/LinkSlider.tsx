import { createSignal } from 'solid-js'
import { patchControls } from '../lib/api'

const debounceMs = 50
const sliderMax = 50
const percentDivisor = 100

export default function LinkSlider(props: { id: string }) {
  const [loss, setLoss] = createSignal(0)
  let pending: ReturnType<typeof setTimeout> | undefined

  function onChange(value: number) {
    setLoss(value)
    clearTimeout(pending)
    pending = setTimeout(() => {
      patchControls(props.id, { loss: value / percentDivisor })
    }, debounceMs)
  }

  return (
    <div class="mt-2">
      <label class="text-xs text-muted">
        loss: <span class="font-mono text-accent">{loss().toFixed(0)}%</span>
      </label>
      <input
        type="range"
        min="0"
        max={sliderMax}
        step="1"
        value={loss()}
        onInput={(e) => onChange(parseInt(e.currentTarget.value, 10))}
        class="w-full"
      />
    </div>
  )
}
