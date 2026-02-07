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

            // Initialize search index with prioritized fields
            this.searchIndex = new FlexSearch.Document({
                document: {
                    id: 'id',
                    index: [
                        { field: 'title', tokenize: 'forward', optimize: true, resolution: 9 },
                        { field: 'content', tokenize: 'forward', optimize: true, resolution: 5 }
                    ],
                    store: true
                }
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
                        this.$nextTick(() => {
                            this.$refs.themeSearch?.focus();
                            this.scrollToCurrentTheme();
                        });
                        break;
                    case 'N':
                        this.lineNumbers = !this.lineNumbers;
                        break;
                    case ',':
                    case '<':
                        if (e.shiftKey && e.altKey) {
                            this.currentSlide = 0;
                            this.updateHash();
                        }
                        break;
                    case '.':
                    case '>':
                        if (e.shiftKey && e.altKey) {
                            this.currentSlide = this.totalSlides - 1;
                            this.updateHash();
                        }
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

            this.$watch('searchQuery', () => {
                if (!this.searchQuery) {
                    this.$nextTick(() => this.scrollToCurrentTheme());
                }
            });
        },

        scrollToCurrentTheme() {
            const el = document.querySelector(`[data-theme-name="${this.currentTheme}"]`);
            if (el) el.scrollIntoView({ block: 'nearest' });
        },

        performSearch(query) {
            this.searchQuery = query;
            if (!query) {
                this.searchResults = [];
                return;
            }

            const results = this.searchIndex.search(query, { limit: 20, enrich: true });

            // FlexSearch results are grouped by field.
            // We want to prioritize title matches.
            let titleMatches = [];
            let contentMatches = [];

            results.forEach(r => {
                if (r.field === 'title') titleMatches = r.result;
                if (r.field === 'content') contentMatches = r.result;
            });

            // Merge results, prioritizing titles and ensuring uniqueness
            const merged = [...titleMatches];
            contentMatches.forEach(cm => {
                if (!merged.find(m => m.id === cm.id)) {
                    merged.push(cm);
                }
            });

            this.searchResults = merged.map(m => ({ id: m.id, title: m.doc.title }));
        },

        jumpToSlide(id) {
            this.currentSlide = id;
            this.showSearch = false;
            this.searchQuery = '';
            this.searchResults = [];
            this.updateHash();
        },

        searchThemes(query) {
            this.searchQuery = query; // Use same query variable for $watch
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
            this.searchQuery = '';
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
