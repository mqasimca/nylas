// ====================================
// CORE API CLIENT
// Base HTTP client for Nylas Air UI
// With retry logic and rate limit handling
// ====================================

const AirAPI = {
    baseURL: '/api',

    // Rate limiting configuration
    _requestQueue: [],
    _isProcessingQueue: false,
    _minRequestInterval: 200, // Minimum 200ms between requests
    _lastRequestTime: 0,

    // Retry configuration
    _maxRetries: 3,
    _baseDelay: 1000, // 1 second base delay

    // Sleep utility
    _sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    },

    // Queue a request to prevent rate limiting
    async _queueRequest(requestFn) {
        return new Promise((resolve, reject) => {
            this._requestQueue.push({ requestFn, resolve, reject });
            this._processQueue();
        });
    },

    // Process queued requests with rate limiting
    async _processQueue() {
        if (this._isProcessingQueue || this._requestQueue.length === 0) {
            return;
        }

        this._isProcessingQueue = true;

        while (this._requestQueue.length > 0) {
            const { requestFn, resolve, reject } = this._requestQueue.shift();

            // Ensure minimum interval between requests
            const now = Date.now();
            const timeSinceLastRequest = now - this._lastRequestTime;
            if (timeSinceLastRequest < this._minRequestInterval) {
                await this._sleep(this._minRequestInterval - timeSinceLastRequest);
            }

            try {
                this._lastRequestTime = Date.now();
                const result = await requestFn();
                resolve(result);
            } catch (error) {
                reject(error);
            }
        }

        this._isProcessingQueue = false;
    },

    // Generic request method with retry and backoff
    async request(endpoint, options = {}) {
        const requestFn = async () => {
            return this._executeRequest(endpoint, options);
        };

        // Queue the request to prevent overwhelming the API
        return this._queueRequest(requestFn);
    },

    // Execute a single request with retry logic
    async _executeRequest(endpoint, options = {}, retryCount = 0) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);

            // Handle rate limiting (429)
            if (response.status === 429) {
                if (retryCount < this._maxRetries) {
                    const retryAfter = response.headers.get('Retry-After');
                    const delay = retryAfter
                        ? parseInt(retryAfter, 10) * 1000
                        : this._baseDelay * Math.pow(2, retryCount);

                    console.warn(`Rate limited on ${endpoint}, retrying in ${delay}ms...`);
                    await this._sleep(delay);
                    return this._executeRequest(endpoint, options, retryCount + 1);
                }
                throw new Error('Rate limit exceeded. Please try again later.');
            }

            // Handle server errors with retry
            if (response.status >= 500 && retryCount < this._maxRetries) {
                const delay = this._baseDelay * Math.pow(2, retryCount);
                console.warn(`Server error on ${endpoint}, retrying in ${delay}ms...`);
                await this._sleep(delay);
                return this._executeRequest(endpoint, options, retryCount + 1);
            }

            // Parse response
            let data;
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await response.json();
            } else {
                data = { message: await response.text() };
            }

            if (!response.ok) {
                throw new Error(data.error || `HTTP ${response.status}`);
            }

            return data;
        } catch (error) {
            // Retry on network errors
            if (error.name === 'TypeError' && retryCount < this._maxRetries) {
                const delay = this._baseDelay * Math.pow(2, retryCount);
                console.warn(`Network error on ${endpoint}, retrying in ${delay}ms...`);
                await this._sleep(delay);
                return this._executeRequest(endpoint, options, retryCount + 1);
            }

            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    }
};
