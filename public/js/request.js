// request.js - Universal Request Library

const API_BASE = '/api';

/**
 * Global Request Wrapper
 * @param {string} url 
 * @param {object} options 
 * @returns {Promise<any>}
 */
async function request(url, options = {}) {
    const defaultHeaders = {
        'Content-Type': 'application/json'
    };

    const config = {
        method: 'GET',
        headers: defaultHeaders,
        ...options
    };

    if (config.body && typeof config.body === 'object') {
        config.body = JSON.stringify(config.body);
    }

    try {
        const res = await fetch(`${API_BASE}${url}`, config);
        
        // Handle non-JSON responses (e.g. CSV export)
        const contentType = res.headers.get('content-type');
        if (contentType && !contentType.includes('application/json')) {
            return res; // Return raw response for special handling
        }

        const data = await res.json();

        // Handle standard response structure { code: 0, message: "success", data: ... }
        // If the API returns the new structure
        if ('code' in data) {
            if (data.code !== 0) {
                throw new Error(data.message || 'Unknown Error');
            }
            return data.data; // Return the actual data
        }

        // Fallback for legacy APIs or raw returns (if any left)
        if (!res.ok) {
            throw new Error(data.error || data.message || res.statusText);
        }

        return data;

    } catch (error) {
        console.error('Request Error:', error);
        if (window.showToast) {
            showToast(error.message, 'error');
        } else {
            alert(error.message); // Fallback
        }
        throw error;
    }
}

// Export for module usage (if using ES modules) or global
window.request = request;
window.API_BASE = API_BASE; // Keep compatibility
