import './vendor/flexsearch.min.js';
import './vendor/alpine.min.js';

document.addEventListener('alpine:init', () => {
    Alpine.data('slideshow', () => ({
        currentSlide: 0,
        totalSlides: 0,
        showThemePalette: false,
        showFontThemePalette: false,
        showSearch: false,
        showHelp: false,
        showPause: false,
        lineNumbers: true,
        selectedPaletteIndex: 0,

        // Search
        searchQuery: '',
        themeQuery: '',
        fontThemeQuery: '',
        searchResults: [],
        allSlides: [],
        searchIndex: null,

        // Timer
        timerRunning: false,
        pauseMessage: 'Taking a break...',
        pauseMinutes: 5,
        timeRemaining: '00:00',
        endTime: null,
        timerInterval: null,

        // Color Themes
        themes: [
            "light", "dark", "cupcake", "bumblebee", "emerald", "corporate", "synthwave", "retro", "cyberpunk", "valentine", "halloween", "garden", "forest", "aqua", "lofi", "pastel", "fantasy", "wireframe", "black", "luxury", "dracula", "cmyk", "autumn", "business", "acid", "lemonade", "night", "coffee", "winter", "dim", "nord", "sunset", "caramellatte", "abyss", "silk"
        ],
        filteredThemes: [],
        currentTheme: '',

        // Font Themes
        fontThemes: [
            { name: "modern", displayName: "Modern", description: "Clean sans-serif for professional presentations", category: "sans", headingFont: "'Inter', sans-serif", bodyFont: "'Inter', sans-serif", codeFont: "'JetBrains Mono', monospace", googleFonts: ["Inter:wght@400;500;600;700;800;900", "JetBrains+Mono:wght@400;500"] },
            { name: "elegant", displayName: "Elegant", description: "Sophisticated serif headings with clean body", category: "mixed", headingFont: "'Playfair Display', serif", bodyFont: "'Source Sans 3', sans-serif", codeFont: "'Fira Code', monospace", googleFonts: ["Playfair+Display:wght@400;500;600;700;800;900", "Source+Sans+3:wght@400;600;700", "Fira+Code:wght@400;500"] },
            { name: "classic", displayName: "Classic", description: "Timeless serif for formal presentations", category: "serif", headingFont: "'Libre Baskerville', serif", bodyFont: "'Libre Baskerville', serif", codeFont: "'Source Code Pro', monospace", googleFonts: ["Libre+Baskerville:wght@400;700", "Source+Code+Pro:wght@400;500"] },
            { name: "tech", displayName: "Tech", description: "Monospace-inspired technical aesthetic", category: "mono", headingFont: "'Space Grotesk', sans-serif", bodyFont: "'Space Grotesk', sans-serif", codeFont: "'JetBrains Mono', monospace", googleFonts: ["Space+Grotesk:wght@400;500;600;700", "JetBrains+Mono:wght@400;500"] },
            { name: "minimal", displayName: "Minimal", description: "Ultra-clean geometric sans-serif", category: "sans", headingFont: "'DM Sans', sans-serif", bodyFont: "'DM Sans', sans-serif", codeFont: "'DM Mono', monospace", googleFonts: ["DM+Sans:wght@400;500;700", "DM+Mono:wght@400;500"] },
            { name: "editorial", displayName: "Editorial", description: "Magazine-style high contrast", category: "mixed", headingFont: "'Oswald', sans-serif", bodyFont: "'Merriweather', serif", codeFont: "'Roboto Mono', monospace", googleFonts: ["Oswald:wght@400;500;600;700", "Merriweather:wght@400;700", "Roboto+Mono:wght@400;500"] },
            { name: "friendly", displayName: "Friendly", description: "Warm and approachable rounded fonts", category: "sans", headingFont: "'Nunito', sans-serif", bodyFont: "'Nunito', sans-serif", codeFont: "'IBM Plex Mono', monospace", googleFonts: ["Nunito:wght@400;600;700;800", "IBM+Plex+Mono:wght@400;500"] },
            { name: "corporate", displayName: "Corporate", description: "Professional business presentation style", category: "sans", headingFont: "'Work Sans', sans-serif", bodyFont: "'Work Sans', sans-serif", codeFont: "'Ubuntu Mono', monospace", googleFonts: ["Work+Sans:wght@400;500;600;700", "Ubuntu+Mono:wght@400;700"] },
            { name: "academic", displayName: "Academic", description: "Scholarly serif for research talks", category: "serif", headingFont: "'Crimson Text', serif", bodyFont: "'Crimson Text', serif", codeFont: "'Inconsolata', monospace", googleFonts: ["Crimson+Text:wght@400;600;700", "Inconsolata:wght@400;500"] },
            { name: "creative", displayName: "Creative", description: "Artistic and expressive typography", category: "mixed", headingFont: "'Bebas Neue', sans-serif", bodyFont: "'Lato', sans-serif", codeFont: "'JetBrains Mono', monospace", googleFonts: ["Bebas+Neue", "Lato:wght@400;700", "JetBrains+Mono:wght@400;500"] },
            { name: "luxury", displayName: "Luxury", description: "High-end elegant typography", category: "mixed", headingFont: "'Cormorant Garamond', serif", bodyFont: "'Montserrat', sans-serif", codeFont: "'Fira Code', monospace", googleFonts: ["Cormorant+Garamond:wght@400;500;600;700", "Montserrat:wght@400;500;600", "Fira+Code:wght@400;500"] },
            { name: "startup", displayName: "Startup", description: "Contemporary tech startup aesthetic", category: "sans", headingFont: "'Manrope', sans-serif", bodyFont: "'Manrope', sans-serif", codeFont: "'JetBrains Mono', monospace", googleFonts: ["Manrope:wght@400;500;600;700;800", "JetBrains+Mono:wght@400;500"] },
            { name: "vintage", displayName: "Vintage", description: "Retro-inspired classic typography", category: "mixed", headingFont: "'Abril Fatface', serif", bodyFont: "'Lora', serif", codeFont: "'Courier Prime', monospace", googleFonts: ["Abril+Fatface", "Lora:wght@400;500;600", "Courier+Prime:wght@400;700"] },
            { name: "swiss", displayName: "Swiss", description: "International style clean typography", category: "sans", headingFont: "'Helvetica Neue', Helvetica, Arial, sans-serif", bodyFont: "'Helvetica Neue', Helvetica, Arial, sans-serif", codeFont: "'SF Mono', Monaco, monospace", googleFonts: [] },
            { name: "playful", displayName: "Playful", description: "Fun and energetic for casual talks", category: "mixed", headingFont: "'Fredoka', sans-serif", bodyFont: "'Quicksand', sans-serif", codeFont: "'Nova Mono', monospace", googleFonts: ["Fredoka:wght@400;500;600;700", "Quicksand:wght@400;500;600;700", "Nova+Mono"] },
            { name: "brutalist", displayName: "Brutalist", description: "Bold raw typography with contrast", category: "mixed", headingFont: "'Archivo Black', sans-serif", bodyFont: "'Space Grotesk', sans-serif", codeFont: "'Share Tech Mono', monospace", googleFonts: ["Archivo+Black", "Space+Grotesk:wght@400;500;600", "Share+Tech+Mono"] }
        ],
        filteredFontThemes: [],
        currentFontTheme: '',
        loadedFonts: new Set(),

        refreshSlides() {
            this.allSlides = [];
            const slides = document.querySelectorAll('.slide');
            slides.forEach((slide, i) => {
                const title = slide.querySelector('h1, h2, h3')?.textContent || `Slide ${i+1}`;
                const content = slide.textContent || '';
                const preview = content.trim().substring(0, 100).replace(/\s+/g, ' ') + (content.length > 100 ? '...' : '');
                const slideData = { id: i, title, content: preview };
                this.allSlides.push(slideData);
                if (this.searchIndex) {
                    this.searchIndex.add({ id: i, title, content });
                }
            });
            return this.allSlides;
        },

        init() {
            console.log('Slideshow initialized');
            this.totalSlides = document.querySelectorAll('.slide').length;
            this.currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
            this.originalTheme = this.currentTheme;
            this.filteredThemes = this.themes;

            // Initialize font theme
            this.currentFontTheme = document.documentElement.getAttribute('data-font-theme') || 'modern';
            this.originalFontTheme = this.currentFontTheme;
            this.filteredFontThemes = this.fontThemes;
            this.loadGoogleFonts(this.currentFontTheme);

            // Initialize search index with prioritized fields
            try {
                // Try multiple ways to find FlexSearch constructor
                const FS = (typeof FlexSearch !== 'undefined') ? FlexSearch : (window.FlexSearch || null);

                if (FS) {
                    this.searchIndex = new FS.Document({
                        document: {
                            id: 'id',
                            index: [
                                { field: 'title', tokenize: 'forward', optimize: true, resolution: 9 },
                                { field: 'content', tokenize: 'forward', optimize: true, resolution: 5 }
                            ],
                            store: true
                        }
                    });
                } else {
                    console.warn('FlexSearch not found, search will be disabled');
                }
            } catch (e) {
                console.error('Failed to initialize search index:', e);
            }

            // Initial slide collection
            this.refreshSlides();

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
                if (this.showFontThemePalette && this.filteredFontThemes[index]) {
                    const themeName = this.filteredFontThemes[index].name;
                    document.documentElement.setAttribute('data-font-theme', themeName);
                    this.loadGoogleFonts(themeName);
                    this.$nextTick(() => {
                        const items = this.$refs.fontThemeList?.querySelectorAll('.font-theme-item');
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
                    if (this.showFontThemePalette) {
                        document.documentElement.setAttribute('data-font-theme', this.originalFontTheme);
                        this.currentFontTheme = this.originalFontTheme;
                        this.loadGoogleFonts(this.originalFontTheme);
                    }
                    this.showSearch = false;
                    this.showThemePalette = false;
                    this.showFontThemePalette = false;
                    this.showHelp = false;
                    this.selectedPaletteIndex = 0;
                    this.searchQuery = '';
                    this.themeQuery = '';
                    this.fontThemeQuery = '';
                    return;
                }

                if (this.showFontThemePalette) {
                    const items = this.filteredFontThemes;
                    if (items.length > 0) {
                        if (e.key === 'ArrowDown') {
                            e.preventDefault();
                            this.selectedPaletteIndex = (this.selectedPaletteIndex + 1) % items.length;
                        } else if (e.key === 'ArrowUp') {
                            e.preventDefault();
                            this.selectedPaletteIndex = (this.selectedPaletteIndex - 1 + items.length) % items.length;
                        } else if (e.key === 'Enter') {
                            e.preventDefault();
                            this.setFontTheme(items[this.selectedPaletteIndex].name);
                        }
                    }
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
                        this.themeQuery = '';
                        this.filteredThemes = this.themes;
                        this.showThemePalette = true;
                        this.selectedPaletteIndex = 0;
                        this.$nextTick(() => {
                            this.$refs.themeSearch?.focus();
                            this.scrollToCurrentTheme();
                        });
                        break;
                    case 'T':
                        this.originalFontTheme = this.currentFontTheme;
                        this.fontThemeQuery = '';
                        this.filteredFontThemes = this.fontThemes;
                        this.showFontThemePalette = true;
                        this.selectedPaletteIndex = 0;
                        this.$nextTick(() => {
                            this.$refs.fontThemeSearch?.focus();
                            this.scrollToCurrentFontTheme();
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
                        if (this.allSlides.length === 0) this.refreshSlides();
                        this.searchQuery = '';
                        this.searchResults = [...this.allSlides];
                        this.showSearch = true;
                        this.selectedPaletteIndex = 0;
                        this.$nextTick(() => this.$refs.slideSearch?.focus());
                        break;
                    case 'P':
                        if (e.shiftKey) this.showPause = !this.showPause;
                        break;
                }
            });

            this.$watch('themeQuery', () => {
                if (!this.themeQuery && this.showThemePalette) {
                    this.$nextTick(() => this.scrollToCurrentTheme());
                }
            });
        },

        scrollToCurrentTheme() {
            const el = document.querySelector(`[data-theme-name="${this.currentTheme}"]`);
            if (el) el.scrollIntoView({ block: 'nearest' });
        },

        async performSearch(query) {
            const currentQuery = query;
            this.searchQuery = query;

            if (!query || !query.trim()) {
                this.searchResults = [...this.allSlides];
                this.selectedPaletteIndex = 0;
                return;
            }

            // Split query into tokens (orderless-style: multiple words/fragments)
            const tokens = query.toLowerCase().split(/\s+/).filter(t => t.length > 0);
            if (tokens.length === 0) {
                this.searchResults = [...this.allSlides];
                this.selectedPaletteIndex = 0;
                return;
            }

            // Helper function to check if all tokens match in a given text
            const matchesAllTokens = (text) => {
                const lowerText = text.toLowerCase();
                return tokens.every(token => lowerText.includes(token));
            };

            // Helper to score a match (higher = better)
            // Title matches score higher, more token matches in title = even higher
            const scoreMatch = (slide) => {
                let score = 0;
                const titleLower = slide.title.toLowerCase();
                const contentLower = slide.content.toLowerCase();

                // Check each token
                for (const token of tokens) {
                    // Title match is worth 10 points per token
                    if (titleLower.includes(token)) {
                        score += 10;
                        // Bonus if token appears at start of title
                        if (titleLower.startsWith(token)) score += 5;
                    }
                    // Content match is worth 1 point per token
                    if (contentLower.includes(token)) {
                        score += 1;
                    }
                }

                return score;
            };

            // Simple fallback if search index is not initialized
            if (!this.searchIndex) {
                // Filter slides where ALL tokens match (orderless)
                const matchingSlides = this.allSlides.filter(s =>
                    matchesAllTokens(s.title) || matchesAllTokens(s.content)
                );

                // Sort by score (title matches first)
                matchingSlides.sort((a, b) => scoreMatch(b) - scoreMatch(a));

                this.searchResults = matchingSlides;
                this.selectedPaletteIndex = 0;
                return;
            }

            try {
                // For orderless search, we search with the full query first
                // to get candidates, then filter by requiring all tokens
                const [titleRes, contentRes] = await Promise.all([
                    this.searchIndex.searchAsync(query, { index: 'title', enrich: true, limit: 50 }),
                    this.searchIndex.searchAsync(query, { index: 'content', enrich: true, limit: 50 })
                ]);

                if (currentQuery !== this.searchQuery) return;

                const extract = (res) => {
                    if (!res) return [];
                    if (Array.isArray(res)) {
                        if (res.length > 0 && res[0].result) return res[0].result;
                        return res;
                    }
                    return [];
                };

                const titleMatches = extract(titleRes);
                const contentMatches = extract(contentRes);

                const candidateIds = new Set();
                titleMatches.forEach(item => {
                    const id = (typeof item === 'object') ? item.id : item;
                    if (id !== undefined) candidateIds.add(id);
                });
                contentMatches.forEach(item => {
                    const id = (typeof item === 'object') ? item.id : item;
                    if (id !== undefined) candidateIds.add(id);
                });

                // Process candidates: require all tokens to match (orderless)
                const scoredResults = [];

                for (const id of candidateIds) {
                    const slide = this.allSlides[id];
                    if (!slide) continue;

                    // Orderless check: ALL tokens must match somewhere
                    const titleMatchesAll = matchesAllTokens(slide.title);
                    const contentMatchesAll = matchesAllTokens(slide.content);

                    if (titleMatchesAll || contentMatchesAll) {
                        const score = scoreMatch(slide);
                        scoredResults.push({ id, slide, score });
                    }
                }

                // Also check slides that weren't in FlexSearch results
                // (FlexSearch might miss some matches with its tokenization)
                for (let i = 0; i < this.allSlides.length; i++) {
                    if (candidateIds.has(i)) continue;
                    const slide = this.allSlides[i];

                    const titleMatchesAll = matchesAllTokens(slide.title);
                    const contentMatchesAll = matchesAllTokens(slide.content);

                    if (titleMatchesAll || contentMatchesAll) {
                        const score = scoreMatch(slide);
                        scoredResults.push({ id: i, slide, score });
                    }
                }

                // Sort by score descending
                scoredResults.sort((a, b) => b.score - a.score);

                // Build final results
                const finalResults = scoredResults.map(({ id, slide }) => {
                    const rawContent = slide.content || '';
                    let preview = rawContent;
                    if (rawContent.length > 0) {
                        const stripped = rawContent.trim().substring(0, 100).replace(/\s+/g, ' ');
                        preview = stripped + (rawContent.length > 100 ? '...' : '');
                    }

                    return {
                        id,
                        title: slide.title || `Slide ${id + 1}`,
                        content: preview
                    };
                });

                this.searchResults = finalResults;
                this.selectedPaletteIndex = 0;
            } catch (e) {
                console.error('Search failed:', e);
                if (currentQuery !== this.searchQuery) return;

                // Fallback to simple orderless matching
                const matchingSlides = this.allSlides.filter(s =>
                    matchesAllTokens(s.title) || matchesAllTokens(s.content)
                );
                matchingSlides.sort((a, b) => scoreMatch(b) - scoreMatch(a));

                this.searchResults = matchingSlides;
                this.selectedPaletteIndex = 0;
            }
        },

        jumpToSlide(id) {
            this.currentSlide = id;
            this.showSearch = false;
            this.searchQuery = '';
            this.searchResults = [];
            this.updateHash();
        },

        searchThemes(query) {
            this.themeQuery = query;
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
            this.themeQuery = '';
        },

        // Font Theme methods
        loadGoogleFonts(themeName) {
            const theme = this.fontThemes.find(t => t.name === themeName);
            if (!theme || !theme.googleFonts || theme.googleFonts.length === 0) return;

            // Check if already loaded
            if (this.loadedFonts.has(themeName)) return;

            // Build Google Fonts URL
            const fontFamilies = theme.googleFonts.join('&family=');
            const linkId = `gf-${themeName}`;

            // Check if link already exists
            if (document.getElementById(linkId)) return;

            const link = document.createElement('link');
            link.id = linkId;
            link.rel = 'stylesheet';
            link.href = `https://fonts.googleapis.com/css2?family=${fontFamilies}&display=swap`;

            link.onload = () => {
                this.loadedFonts.add(themeName);
                console.log(`Loaded fonts for ${themeName}`);
            };

            document.head.appendChild(link);
        },

        searchFontThemes(query) {
            this.fontThemeQuery = query;
            this.selectedPaletteIndex = 0;
            if (!query) {
                this.filteredFontThemes = this.fontThemes;
                // Preview the current theme when clearing search
                document.documentElement.setAttribute('data-font-theme', this.currentFontTheme);
                this.loadGoogleFonts(this.currentFontTheme);
                return;
            }

            const searchTerm = query.toLowerCase();
            this.filteredFontThemes = this.fontThemes.filter(t => {
                return t.name.includes(searchTerm) ||
                       t.displayName.toLowerCase().includes(searchTerm) ||
                       t.category.includes(searchTerm);
            });

            // Preview the first match if any
            if (this.filteredFontThemes.length > 0) {
                const firstTheme = this.filteredFontThemes[0].name;
                document.documentElement.setAttribute('data-font-theme', firstTheme);
                this.loadGoogleFonts(firstTheme);
            }
        },

        setFontTheme(themeName) {
            this.currentFontTheme = themeName;
            document.documentElement.setAttribute('data-font-theme', themeName);
            this.loadGoogleFonts(themeName);
            this.showFontThemePalette = false;
            this.fontThemeQuery = '';
        },

        scrollToCurrentFontTheme() {
            const el = document.querySelector(`[data-font-theme-name="${this.currentFontTheme}"]`);
            if (el) el.scrollIntoView({ block: 'nearest' });
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
