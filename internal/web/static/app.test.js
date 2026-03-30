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
