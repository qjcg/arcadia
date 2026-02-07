import './vendor/flexsearch.min.js';
import './vendor/alpine.min.js';

// Setup search index
// FlexSearch will be available globally if imported this way from minified bundle
// or I can use standard ESM if I had it.

document.addEventListener('alpine:init', () => {
    Alpine.data('slideshow', () => ({
        currentSlide: 0,
        totalSlides: 0,
        showThemePalette: false,
        showSearch: false,
        showHelp: false,
        showPause: false,
        lineNumbers: true,

        // Search
        searchQuery: '',
        searchResults: [],
        searchIndex: null,

        // Timer
        timerRunning: false,
        pauseMessage: 'Taking a break...',
        pauseMinutes: 5,
        timeRemaining: '00:00',
        endTime: null,
        timerInterval: null,

        // Themes
        themes: [
            "light", "dark", "cupcake", "bumblebee", "emerald", "corporate", "synthwave", "retro", "cyberpunk", "valentine", "halloween", "garden", "forest", "aqua", "lofi", "pastel", "fantasy", "wireframe", "black", "luxury", "dracula", "cmyk", "autumn", "business", "acid", "lemonade", "night", "coffee", "winter", "dim", "nord", "sunset", "caramellatte", "abyss", "silk"
        ],
        filteredThemes: [],
        currentTheme: '',

        init() {
            this.totalSlides = document.querySelectorAll('.slide').length;
            this.currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
            this.filteredThemes = this.themes;

            // Initialize search index
            this.searchIndex = new FlexSearch.Document({
                document: {
                    id: 'id',
                    index: ['title', 'content'],
                    store: true
                },
                tokenize: 'forward'
            });

            document.querySelectorAll('.slide').forEach((slide, i) => {
                const title = slide.querySelector('h1, h2, h3')?.textContent || `Slide ${i+1}`;
                const content = slide.textContent;
                this.searchIndex.add({ id: i, title, content });
            });

            // Load persistent timer
            const savedTimer = localStorage.getItem('slidesdeck_timer');
            if (savedTimer) {
                const data = JSON.parse(savedTimer);
                if (new Date(data.endTime) > new Date()) {
                    this.pauseMessage = data.message;
                    this.endTime = new Date(data.endTime);
                    this.showPause = true;
                    this.startCountdown();
                }
            }

            // Handle hash for direct links to slides
            const hash = window.location.hash;
            if (hash) {
                const index = parseInt(hash.substring(1));
                if (!isNaN(index)) this.currentSlide = index;
            }

            window.addEventListener('keydown', (e) => {
                if (this.showSearch || this.showThemePalette) return;

                switch(e.key) {
                    case 'n':
                    case 'ArrowRight':
                    case ' ':
                        if (!this.showPause) this.next();
                        break;
                    case 'p':
                    case 'ArrowLeft':
                        if (!this.showPause) this.prev();
                        break;
                    case 'f':
                        this.toggleFullscreen();
                        break;
                    case 't':
                        this.showThemePalette = true;
                        this.$nextTick(() => this.$refs.themeSearch?.focus());
                        break;
                    case 'N':
                        this.lineNumbers = !this.lineNumbers;
                        break;
                    case '?':
                        this.showHelp = !this.showHelp;
                        break;
                    case '/':
                        e.preventDefault();
                        this.showSearch = true;
                        this.$nextTick(() => this.$refs.slideSearch?.focus());
                        break;
                    case 'P':
                        if (e.shiftKey) this.showPause = !this.showPause;
                        break;
                }
            });
        },

        performSearch(query) {
            this.searchQuery = query;
            if (!query) {
                this.searchResults = [];
                return;
            }
            const results = this.searchIndex.search(query, { limit: 10, enrich: true });
            // Combine and sort results
            let combined = [];
            if (results.length > 0) {
                results.forEach(r => {
                    r.result.forEach(item => {
                        if (!combined.find(c => c.id === item.id)) {
                            combined.push({ id: item.id, title: item.doc.title });
                        }
                    });
                });
            }
            this.searchResults = combined;
        },

        jumpToSlide(id) {
            this.currentSlide = id;
            this.showSearch = false;
            this.searchQuery = '';
            this.searchResults = [];
            this.updateHash();
        },

        searchThemes(query) {
            if (!query) {
                this.filteredThemes = this.themes;
                return;
            }

            const isDarkSearch = query.startsWith('dark:');
            const isLightSearch = query.startsWith('light:');

            let searchTerm = query;
            if (isDarkSearch) searchTerm = query.substring(5).trim();
            if (isLightSearch) searchTerm = query.substring(6).trim();

            this.filteredThemes = this.themes.filter(t => {
                const matchesSearch = t.includes(searchTerm.toLowerCase());
                if (!matchesSearch) return false;

                // Heuristic for light/dark (daisyui names are usually descriptive)
                // In a real version we'd use a map.
                const isDark = ['dark', 'synthwave', 'halloween', 'forest', 'aqua', 'black', 'luxury', 'dracula', 'business', 'night', 'coffee', 'dim', 'sunset', 'abyss'].includes(t);

                if (isDarkSearch) return isDark;
                if (isLightSearch) return !isDark;
                return true;
            });
        },

        setTheme(theme) {
            this.currentTheme = theme;
            document.documentElement.setAttribute('data-theme', theme);
            this.showThemePalette = false;
        },

        startTimer() {
            const now = new Date();
            this.endTime = new Date(now.getTime() + this.pauseMinutes * 60000);
            this.timerRunning = true;
            localStorage.setItem('slidesdeck_timer', JSON.stringify({
                endTime: this.endTime,
                message: this.pauseMessage
            }));
            this.startCountdown();
        },

        startCountdown() {
            this.timerRunning = true;
            this.updateRemaining();
            this.timerInterval = setInterval(() => this.updateRemaining(), 1000);
        },

        updateRemaining() {
            const now = new Date();
            const diff = this.endTime - now;
            if (diff <= 0) {
                this.timeRemaining = '00:00';
                this.stopTimer();
                return;
            }
            const mins = Math.floor(diff / 60000);
            const secs = Math.floor((diff % 60000) / 1000);
            this.timeRemaining = `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        },

        stopTimer() {
            clearInterval(this.timerInterval);
            this.timerRunning = false;
        },

        resetTimer() {
            this.stopTimer();
            this.timeRemaining = '00:00';
            localStorage.removeItem('slidesdeck_timer');
        },

        next() {
            if (this.currentSlide < this.totalSlides - 1) {
                this.currentSlide++;
                this.updateHash();
            }
        },

        prev() {
            if (this.currentSlide > 0) {
                this.currentSlide--;
                this.updateHash();
            }
        },

        updateHash() {
            window.location.hash = this.currentSlide;
        },

        toggleFullscreen() {
            if (!document.fullscreenElement) {
                document.documentElement.requestFullscreen();
            } else {
                if (document.exitFullscreen) {
                    document.exitFullscreen();
                }
            }
        }
    }));
});
