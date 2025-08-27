// Plausible Analytics Tracking Script
(function() {
    'use strict';

    var endpoint = window.plausible_endpoint || 'http://localhost:8080/api/event';
    var domain = window.location.hostname;

    function trigger(eventName, options) {
        options = options || {};
        
        var payload = {
            name: eventName,
            domain: domain,
            url: window.location.href,
            referrer: document.referrer || '',
            props: options.props || {}
        };

        // Add UTM parameters if present
        var urlParams = new URLSearchParams(window.location.search);
        ['utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content'].forEach(function(param) {
            var value = urlParams.get(param);
            if (value) {
                payload[param] = value;
            }
        });

        // Send the event
        var request = new XMLHttpRequest();
        request.open('POST', endpoint, true);
        request.setRequestHeader('Content-Type', 'application/json');
        
        request.send(JSON.stringify(payload));
    }

    // Track pageview on load
    function trackPageview() {
        trigger('pageview');
    }

    // Track custom events
    function plausible(eventName, options) {
        trigger(eventName, options);
    }

    // Auto-track pageviews
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', trackPageview);
    } else {
        trackPageview();
    }

    // Track SPA navigation
    var originalPushState = history.pushState;
    var originalReplaceState = history.replaceState;

    history.pushState = function() {
        originalPushState.apply(history, arguments);
        setTimeout(trackPageview, 0);
    };

    history.replaceState = function() {
        originalReplaceState.apply(history, arguments);
        setTimeout(trackPageview, 0);
    };

    window.addEventListener('popstate', trackPageview);

    // Expose plausible function globally
    window.plausible = plausible;
})();
