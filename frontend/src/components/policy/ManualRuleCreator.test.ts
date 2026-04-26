import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ManualRuleCreator from './ManualRuleCreator.vue'

const SelectStub = defineComponent({
  props: {
    modelValue: {
      type: [String, Number],
      required: true
    },
    options: {
      type: Array,
      required: true
    }
  },
  emits: ['update:modelValue'],
  template: `
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option v-for="option in options" :key="String(option.value)" :value="option.value">
        {{ option.label }}
      </option>
    </select>
  `
})

function mountCreator(props?: Record<string, unknown>) {
  return mount(ManualRuleCreator, {
    props,
    global: {
      stubs: {
        Select: SelectStub
      }
    }
  })
}

describe('component/ManualRuleCreator', () => {
  beforeEach(() => {
    vi.stubGlobal('confirm', vi.fn(() => true))
  })

  it('requires the match-specific fields needed to enable rule creation', async () => {
    const wrapper = mountCreator()
    const actionButton = wrapper.get('button.btn-primary')

    expect(actionButton.attributes('disabled')).toBeDefined()

    await wrapper.get('input[placeholder="e.g., Block Suspicious Process"]').setValue('Watch Bash')
    await wrapper.get('textarea').setValue('Alert on bash')
    await wrapper.get('input[placeholder="e.g., /usr/bin/bash"]').setValue('/usr/bin/bash')
    expect(actionButton.attributes('disabled')).toBeUndefined()

    const selects = wrapper.findAll('select')
    await selects[3].setValue('connect')
    expect(actionButton.attributes('disabled')).toBeDefined()

    await wrapper.get('input[placeholder="e.g., 3306"]').setValue('3306')
    expect(actionButton.attributes('disabled')).toBeUndefined()
  })

  it('emits created rules with the built match object and generated YAML', async () => {
    const wrapper = mountCreator()

    await wrapper.get('input[placeholder="e.g., Block Suspicious Process"]').setValue('Watch File')
    await wrapper.get('textarea').setValue('Alert on file access')
    await wrapper.findAll('select')[3].setValue('file')
    await wrapper.get('input[placeholder="e.g., /tmp/suspicious.sh"]').setValue('/tmp/suspicious.sh')
    await wrapper.get('button.btn-primary').trigger('click')

    const emitted = wrapper.emitted('rule-created')
    expect(emitted).toHaveLength(1)
    expect(emitted?.[0]?.[0]).toMatchObject({
      name: 'Watch File',
      description: 'Alert on file access',
      type: 'file',
      match: {
        filename: '/tmp/suspicious.sh'
      }
    })
    expect(emitted?.[0]?.[0].yaml).toContain('name: Watch File')
    expect(emitted?.[0]?.[0].yaml).toContain('filename: /tmp/suspicious.sh')
  })

  it('hydrates existing rules, emits updates, and respects delete confirmation', async () => {
    vi.stubGlobal('confirm', vi.fn(() => false))

    const wrapper = mountCreator({
      rule: {
        name: 'Block DB',
        description: 'Block DB connections',
        action: 'block',
        severity: 'critical',
        state: 'production',
        type: 'connect',
        match: {
          destPort: 3306,
          destIp: '192.168.1.100'
        },
        yaml: ''
      }
    })

    await wrapper.get('button.btn-primary').trigger('click')
    expect(wrapper.emitted('rule-updated')?.[0]?.[0]).toMatchObject({
      name: 'Block DB',
      action: 'block',
      state: 'production',
      type: 'connect',
      match: {
        destPort: 3306,
        destIp: '192.168.1.100'
      }
    })

    await wrapper.get('button.btn-danger').trigger('click')
    expect(wrapper.emitted('rule-deleted')).toBeUndefined()

    vi.stubGlobal('confirm', vi.fn(() => true))
    await wrapper.get('button.btn-danger').trigger('click')
    expect(wrapper.emitted('rule-deleted')?.[0]).toEqual(['Block DB'])
  })
})
