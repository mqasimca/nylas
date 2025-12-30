/**
 * App Accessibility - Accessibility helpers
 */
        // ACCESSIBILITY HELPERS
        // NOTE: announce() is in js/utils.js
        // NOTE: Email keyboard navigation is in js/email.js (initEmailKeyboard)
        // ====================================

        // Trap focus within modal (unique to app.js)
        function trapFocus(element) {
            const focusableElements = element.querySelectorAll(
                'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
            );
            const firstFocusable = focusableElements[0];
            const lastFocusable = focusableElements[focusableElements.length - 1];

            element.addEventListener('keydown', function(e) {
                if (e.key === 'Tab') {
                    if (e.shiftKey) {
                        if (document.activeElement === firstFocusable) {
                            lastFocusable.focus();
                            e.preventDefault();
                        }
                    } else {
                        if (document.activeElement === lastFocusable) {
                            firstFocusable.focus();
                            e.preventDefault();
                        }
                    }
                }
            });

            // Focus first element
            firstFocusable?.focus();
        }
