import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Wallpaper {
    id: string;
    name: string;
    type: 'gradient' | 'image';
    value: string; // CSS gradient or image path
    thumbnail?: string;
}

export const predefinedWallpapers: Wallpaper[] = [
    {
        id: 'gradient-default',
        name: 'Default Gradient',
        type: 'gradient',
        value: 'linear-gradient(135deg, hsl(220, 70%, 90%) 0%, hsl(240, 60%, 95%) 50%, hsl(200, 65%, 92%) 100%)'
    },
    {
        id: 'gradient-dark',
        name: 'Dark Gradient',
        type: 'gradient',
        value: 'linear-gradient(135deg, hsl(220, 40%, 8%) 0%, hsl(240, 30%, 12%) 50%, hsl(200, 35%, 10%) 100%)'
    },
    {
        id: 'blue-waves',
        name: 'Blue Waves',
        type: 'image',
        value: '/wallpapers/blue-waves.png',
        thumbnail: '/wallpapers/blue-waves.png'
    },
    {
        id: 'geometric-dark',
        name: 'Geometric Dark',
        type: 'image',
        value: '/wallpapers/geometric-dark.png',
        thumbnail: '/wallpapers/geometric-dark.png'
    }
];

const STORAGE_KEY = 'mynt-wallpaper';

function getInitialWallpaper(): Wallpaper {
    if (browser) {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
            try {
                const parsed = JSON.parse(stored);
                // Verify it's a valid wallpaper
                const found = predefinedWallpapers.find(w => w.id === parsed.id);
                if (found) return found;
            } catch (e) {
                console.error('Failed to parse stored wallpaper:', e);
            }
        }
    }
    // Default to gradient
    return predefinedWallpapers[0];
}

function createWallpaperStore() {
    const { subscribe, set, update } = writable<Wallpaper>(getInitialWallpaper());

    return {
        subscribe,
        set: (wallpaper: Wallpaper) => {
            set(wallpaper);
            if (browser) {
                localStorage.setItem(STORAGE_KEY, JSON.stringify(wallpaper));
            }
        },
        selectById: (id: string) => {
            const wallpaper = predefinedWallpapers.find(w => w.id === id);
            if (wallpaper) {
                set(wallpaper);
                if (browser) {
                    localStorage.setItem(STORAGE_KEY, JSON.stringify(wallpaper));
                }
            }
        }
    };
}

export const currentWallpaper = createWallpaperStore();
