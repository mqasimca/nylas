/**
 * App Effects - Visual effects (parallax, AI typing, magnetic buttons, spring animations)
 */
        // Parallax effect on cards
        document.querySelectorAll('.email-item').forEach(card => {
            card.addEventListener('mousemove', function(e) {
                const rect = card.getBoundingClientRect();
                const x = e.clientX - rect.left;
                const y = e.clientY - rect.top;
                const centerX = rect.width / 2;
                const centerY = rect.height / 2;
                const rotateX = (y - centerY) / 20;
                const rotateY = (centerX - x) / 20;
                card.style.setProperty('--rotateX', rotateX + 'deg');
                card.style.setProperty('--rotateY', rotateY + 'deg');
            });

            card.addEventListener('mouseleave', function() {
                card.style.setProperty('--rotateX', '0deg');
                card.style.setProperty('--rotateY', '0deg');
            });
        });

        // Demo: AI Typing Animation
        function showAITyping(element, text) {
            element.classList.add('ai-streaming-text');
            let i = 0;
            const interval = setInterval(() => {
                element.textContent = text.substring(0, i);
                i++;
                if (i > text.length) {
                    clearInterval(interval);
                    element.classList.remove('ai-streaming-text');
                }
            }, 30);
        }

        // Magnetic button effect
        document.querySelectorAll('.magnetic-btn, .compose-btn, .action-btn').forEach(btn => {
            btn.addEventListener('mousemove', function(e) {
                const rect = btn.getBoundingClientRect();
                const x = e.clientX - rect.left - rect.width / 2;
                const y = e.clientY - rect.top - rect.height / 2;
                btn.style.transform = `translate(${x * 0.1}px, ${y * 0.1}px)`;
            });

            btn.addEventListener('mouseleave', function() {
                btn.style.transform = '';
            });
        });

        // Spring animation on new elements
        function springAnimate(element) {
            element.classList.add('spring-in');
            element.addEventListener('animationend', () => {
                element.classList.remove('spring-in');
            }, { once: true });
        }
