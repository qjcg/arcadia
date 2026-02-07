import './vendor/flexsearch.min.js';
import './vendor/alpine.min.js';

document.addEventListener('alpine:init', () => {
    Alpine.data('slideshow', () => ({
        currentSlide: 0,
        totalSlides: 0,
        showThemePalette: false,
        showSearch: false,
        showHelp: false,
        showPause: false,
        lineNumbers: true,
        selectedPaletteIndex: 0,

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
            console.log('Slideshow initialized');
            this.totalSlides = document.querySelectorAll('.slide').length;
            this.currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
            this.originalTheme = this.currentTheme;
            this.filteredThemes = this.themes;

            // Initialize search index with prioritized fields
            try {
                if (typeof FlexSearch !== 'undefined') {
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
                } else {
                    console.warn('FlexSearch not found, search will be disabled');
                }
            } catch (e) {
                console.error('Failed to initialize search index:', e);
            }

            // Load persistent timer
            try {
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
            } catch (e) {}

            // Handle hash for direct links to slides
            const hash = window.location.hash;
            if (hash) {
                const index = parseInt(hash.substring(1));
                if (!isNaN(index)) this.currentSlide = index;
            }

            this.$watch('selectedPaletteIndex', (index) => {
                if (this.showThemePalette && this.filteredThemes[index]) {
                    document.documentElement.setAttribute('data-theme', this.filteredThemes[index]);
                    this.$nextTick(() => {
                        const items = this.$refs.themeList?.querySelectorAll('.theme-item');
                        if (items && items[index]) items[index].scrollIntoView({ block: 'nearest' });
                    });
                }
                if (this.showSearch && this.searchResults[index]) {
                    this.$nextTick(() => {
                        const items = this.$refs.searchResults?.querySelectorAll('.search-item');
                        if (items && items[index]) items[index].scrollIntoView({ block: 'nearest' });
                    });
                }
            });

            window.addEventListener('keydown', (e) => {
                // Global shortcuts that work even when palettes are open
                if (e.key === 'Escape') {
                    if (this.showThemePalette) {
                        document.documentElement.setAttribute('data-theme', this.originalTheme);
                        this.currentTheme = this.originalTheme;
                    }
                    this.showSearch = false;
                    this.showThemePalette = false;
                    this.showHelp = false;
                    this.selectedPaletteIndex = 0;
                    this.searchQuery = '';
                    return;
                }

                if (this.showSearch || this.showThemePalette) {
                    const items = this.showSearch ? this.searchResults : this.filteredThemes;
                    if (items.length > 0) {
                        if (e.key === 'ArrowDown') {
                            e.preventDefault();
                            this.selectedPaletteIndex = (this.selectedPaletteIndex + 1) % items.length;
                        } else if (e.key === 'ArrowUp') {
                            e.preventDefault();
                            this.selectedPaletteIndex = (this.selectedPaletteIndex - 1 + items.length) % items.length;
                        } else if (e.key === 'Enter') {
                            e.preventDefault();
                            if (this.showSearch) {
                                this.jumpToSlide(items[this.selectedPaletteIndex].id);
                            } else {
                                this.setTheme(items[this.selectedPaletteIndex]);
                            }
                        }
                    }
                    return;
                }

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
                        this.originalTheme = this.currentTheme;
                        this.searchQuery = '';
                        this.filteredThemes = this.themes;
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
                        this.searchQuery = '';
                        this.searchResults = [];
                        this.showSearch = true;
                        this.selectedPaletteIndex = 0;
                        this.$nextTick(() => this.$refs.slideSearch?.focus());
                        break;
                    case 'P':
                        if (e.shiftKey) this.showPause = !this.showPause;
                        break;
                }
            });

            this.$watch('searchQuery', () => {
                if (!this.searchQuery && this.showThemePalette) {
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
            this.selectedPaletteIndex = 0;
            if (!query || !this.searchIndex) {
                this.searchResults = [];
                return;
            }

            const results = this.searchIndex.search(query, { limit: 20, enrich: true });

            let titleMatches = [];
            let contentMatches = [];

            results.forEach(r => {
                if (r.field === 'title') titleMatches = r.result;
                if (r.field === 'content') contentMatches = r.result;
            });

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
            this.searchQuery = query;
            this.selectedPaletteIndex = 0;
            if (!query) {
                this.filteredThemes = this.themes;
                // Preview the current theme when clearing search
                document.documentElement.setAttribute('data-theme', this.currentTheme);
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

                const isDark = ['dark', 'synthwave', 'halloween', 'forest', 'aqua', 'black', 'luxury', 'dracula', 'business', 'night', 'coffee', 'dim', 'sunset', 'abyss'].includes(t);

                if (isDarkSearch) return isDark;
                if (isLightSearch) return !isDark;
                return true;
            });

            // Preview the first match if any
            if (this.filteredThemes.length > 0) {
                document.documentElement.setAttribute('data-theme', this.filteredThemes[0]);
            }
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
            if (this.timerInterval) clearInterval(this.timerInterval);
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
            this.showPause = false;
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
