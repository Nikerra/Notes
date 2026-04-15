const CACHE_NAME = 'notes-v6';
const STATIC_ASSETS = [
    '/static/styles.css?v=6',
    '/static/app.js?v=6',
    '/static/manifest.json',
    '/static/icon-192.png'
];

self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then((cache) => cache.addAll(STATIC_ASSETS))
    );
});

self.addEventListener('fetch', (event) => {
    const url = new URL(event.request.url);
    
    if (url.pathname === '/' || url.pathname === '/index.html') {
        event.respondWith(
            fetch(event.request)
                .catch(() => caches.match('/'))
        );
        return;
    }
    
    if (url.pathname.startsWith('/static/')) {
        event.respondWith(
            caches.match(event.request)
                .then((response) => {
                    if (response) {
                        fetch(event.request)
                            .then((newResponse) => {
                                if (newResponse && newResponse.status === 200) {
                                    caches.open(CACHE_NAME)
                                        .then((cache) => cache.put(event.request, newResponse));
                                }
                            });
                        return response;
                    }
                    return fetch(event.request);
                })
        );
        return;
    }
    
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(
            fetch(event.request)
                .catch(() => new Response(JSON.stringify([]), {
                    headers: { 'Content-Type': 'application/json' }
                }))
        );
        return;
    }
    
    event.respondWith(
        fetch(event.request)
            .then((response) => {
                if (response && response.status === 200 && response.type === 'basic') {
                    const responseToCache = response.clone();
                    caches.open(CACHE_NAME)
                        .then((cache) => cache.put(event.request, responseToCache));
                }
                return response;
            })
            .catch(() => caches.match(event.request))
    );
});

self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    if (cacheName !== CACHE_NAME) {
                        console.log('Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        })
    );
});
