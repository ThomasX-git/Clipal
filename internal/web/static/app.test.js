const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

function loadApp() {
    const source = fs.readFileSync(path.join(__dirname, 'app.js'), 'utf8');
    const context = {
        console,
        localStorage: {
            getItem() {
                return null;
            },
            setItem() {}
        },
        document: {
            documentElement: {
                lang: 'en',
                classList: {
                    add() {},
                    remove() {}
                }
            },
            getElementById() {
                return null;
            }
        },
        window: {
            matchMedia() {
                return {
                    matches: false,
                    addEventListener() {}
                };
            }
        },
        setTimeout,
        clearTimeout,
        queueMicrotask,
        URL,
        fetch: async () => ({ ok: true, json: async () => ({}) }),
        confirm: () => true
    };

    vm.runInNewContext(`${source}\n;globalThis.__appFactory = app;`, context, {
        filename: 'app.js'
    });

    return context.__appFactory();
}

test('applyLocalProviderReorder reorders providers and renumbers priority immediately', () => {
    const state = loadApp();
    state.providers = [
        { name: 'p1', priority: 1, enabled: true },
        { name: 'p2', priority: 2, enabled: true },
        { name: 'p3', priority: 3, enabled: true }
    ];

    state.applyLocalProviderReorder(1, 0);

    assert.deepEqual(
        state.providers.map(provider => `${provider.name}:${provider.priority}`),
        ['p2:1', 'p1:2', 'p3:3']
    );
});

test('syncSortableDomOrder delegates to Sortable instance with provider names', () => {
    const state = loadApp();
    const calls = [];
    state.sortableInstance = {
        sort(order, useAnimation) {
            calls.push({ order, useAnimation });
        }
    };

    state.syncSortableDomOrder(['p3', 'p1', 'p2']);

    assert.deepEqual(calls, [
        { order: ['p3', 'p1', 'p2'], useAnimation: false }
    ]);
});

test('afterProviderRender aligns Sortable DOM order to current providers on nextTick', async () => {
    const state = loadApp();
    const calls = [];
    state.providers = [
        { name: 'p2', priority: 1, enabled: true },
        { name: 'p1', priority: 2, enabled: true }
    ];
    state.sortableInstance = {
        sort(order, useAnimation) {
            calls.push({ order, useAnimation });
        }
    };
    state.$nextTick = callback => callback();

    await state.afterProviderRender();

    assert.deepEqual(calls, [
        { order: ['p2', 'p1'], useAnimation: false }
    ]);
});

test('saveProvider includes OpenAI override fields in payload', async () => {
    const state = loadApp();
    const calls = [];
    state.selectedClient = 'openai';
    state.providerForm = {
        name: 'openai-primary',
        base_url: 'https://example.com',
        model: 'gpt-5.4',
        reasoning_effort: 'high',
        thinking_budget_tokens: 0,
        api_keys_text: 'key-1',
        priority: 1,
        enabled: true
    };
    state.apiCall = async (url, options) => {
        calls.push({ url, options: JSON.parse(options.body) });
        return {};
    };
    state.showAlert = () => {};
    state.closeModals = () => {};
    state.loadProviders = async () => {};
    state.refreshStatus = async () => {};

    await state.saveProvider();

    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, '/api/providers/openai');
    assert.deepEqual(calls[0].options, {
        name: 'openai-primary',
        base_url: 'https://example.com',
        priority: 1,
        enabled: true,
        model: 'gpt-5.4',
        reasoning_effort: 'high',
        api_key: 'key-1'
    });
});

test('saveProvider includes Claude thinking budget override in payload', async () => {
    const state = loadApp();
    const calls = [];
    state.selectedClient = 'claude';
    state.providerForm = {
        name: 'claude-primary',
        base_url: 'https://example.com',
        model: 'claude-sonnet-4-5',
        reasoning_effort: '',
        thinking_budget_tokens: 2048,
        api_keys_text: 'key-1\nkey-2',
        priority: 2,
        enabled: false
    };
    state.apiCall = async (url, options) => {
        calls.push({ url, options: JSON.parse(options.body) });
        return {};
    };
    state.showAlert = () => {};
    state.closeModals = () => {};
    state.loadProviders = async () => {};
    state.refreshStatus = async () => {};

    await state.saveProvider();

    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, '/api/providers/claude');
    assert.deepEqual(calls[0].options, {
        name: 'claude-primary',
        base_url: 'https://example.com',
        priority: 2,
        enabled: false,
        model: 'claude-sonnet-4-5',
        thinking_budget_tokens: 2048,
        api_keys: ['key-1', 'key-2']
    });
});
