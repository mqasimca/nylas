/**
 * Confetti celebration for Inbox Zero
 * Lightweight, performant confetti animation
 */

class Confetti {
    constructor(options = {}) {
        this.canvas = null;
        this.ctx = null;
        this.particles = [];
        this.animationId = null;
        this.isRunning = false;

        this.options = {
            particleCount: options.particleCount || 100,
            spread: options.spread || 70,
            startVelocity: options.startVelocity || 30,
            decay: options.decay || 0.95,
            gravity: options.gravity || 1,
            drift: options.drift || 0,
            ticks: options.ticks || 200,
            colors: options.colors || [
                '#6366f1', // Indigo
                '#8b5cf6', // Violet
                '#ec4899', // Pink
                '#06b6d4', // Cyan
                '#10b981', // Emerald
                '#f59e0b', // Amber
            ],
            shapes: options.shapes || ['square', 'circle'],
            scalar: options.scalar || 1,
            origin: options.origin || { x: 0.5, y: 0.5 }
        };
    }

    createCanvas() {
        if (this.canvas) return;

        this.canvas = document.createElement('canvas');
        this.canvas.className = 'confetti-canvas';
        this.canvas.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: 9999;
        `;
        document.body.appendChild(this.canvas);
        this.ctx = this.canvas.getContext('2d');
        this.resizeCanvas();

        window.addEventListener('resize', () => this.resizeCanvas());
    }

    resizeCanvas() {
        if (!this.canvas) return;
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }

    createParticle(x, y) {
        const angle = Math.random() * Math.PI * 2;
        const velocity = this.options.startVelocity * (0.5 + Math.random() * 0.5);
        const spread = this.options.spread * (Math.PI / 180);

        return {
            x,
            y,
            vx: Math.cos(angle - spread + Math.random() * spread * 2) * velocity,
            vy: Math.sin(angle - spread + Math.random() * spread * 2) * velocity - 10,
            color: this.options.colors[Math.floor(Math.random() * this.options.colors.length)],
            shape: this.options.shapes[Math.floor(Math.random() * this.options.shapes.length)],
            size: (5 + Math.random() * 5) * this.options.scalar,
            rotation: Math.random() * Math.PI * 2,
            rotationSpeed: (Math.random() - 0.5) * 0.2,
            ticks: this.options.ticks,
            opacity: 1
        };
    }

    drawParticle(particle) {
        const { ctx } = this;

        ctx.save();
        ctx.translate(particle.x, particle.y);
        ctx.rotate(particle.rotation);
        ctx.globalAlpha = particle.opacity;
        ctx.fillStyle = particle.color;

        if (particle.shape === 'circle') {
            ctx.beginPath();
            ctx.arc(0, 0, particle.size / 2, 0, Math.PI * 2);
            ctx.fill();
        } else {
            ctx.fillRect(-particle.size / 2, -particle.size / 2, particle.size, particle.size);
        }

        ctx.restore();
    }

    updateParticle(particle) {
        particle.x += particle.vx;
        particle.y += particle.vy;
        particle.vx *= this.options.decay;
        particle.vy *= this.options.decay;
        particle.vy += this.options.gravity;
        particle.vx += this.options.drift;
        particle.rotation += particle.rotationSpeed;
        particle.ticks--;
        particle.opacity = Math.max(0, particle.ticks / this.options.ticks);

        return particle.ticks > 0;
    }

    animate() {
        if (!this.isRunning) return;

        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        this.particles = this.particles.filter(particle => {
            this.drawParticle(particle);
            return this.updateParticle(particle);
        });

        if (this.particles.length > 0) {
            this.animationId = requestAnimationFrame(() => this.animate());
        } else {
            this.stop();
        }
    }

    fire(options = {}) {
        this.createCanvas();

        const mergedOptions = { ...this.options, ...options };
        const x = mergedOptions.origin.x * this.canvas.width;
        const y = mergedOptions.origin.y * this.canvas.height;

        for (let i = 0; i < mergedOptions.particleCount; i++) {
            this.particles.push(this.createParticle(x, y));
        }

        if (!this.isRunning) {
            this.isRunning = true;
            this.animate();
        }
    }

    burst(options = {}) {
        // Burst from center
        this.fire({
            ...options,
            origin: { x: 0.5, y: 0.5 },
            particleCount: 80,
            startVelocity: 45,
            spread: 360
        });
    }

    cannon(options = {}) {
        // Fire from bottom sides
        this.fire({
            ...options,
            origin: { x: 0.1, y: 0.9 },
            particleCount: 50,
            spread: 55,
            startVelocity: 55
        });

        this.fire({
            ...options,
            origin: { x: 0.9, y: 0.9 },
            particleCount: 50,
            spread: 55,
            startVelocity: 55
        });
    }

    celebration() {
        // Full celebration sequence
        const duration = 3000;
        const interval = 250;
        const end = Date.now() + duration;

        const frame = () => {
            this.fire({
                particleCount: 3,
                spread: 55,
                origin: { x: Math.random(), y: Math.random() * 0.3 }
            });

            if (Date.now() < end) {
                setTimeout(frame, interval);
            }
        };

        // Initial burst
        this.burst();

        // Continue with smaller bursts
        setTimeout(frame, 250);
    }

    stop() {
        this.isRunning = false;
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
            this.animationId = null;
        }
        this.particles = [];
        if (this.ctx) {
            this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        }
    }

    destroy() {
        this.stop();
        if (this.canvas && this.canvas.parentNode) {
            this.canvas.parentNode.removeChild(this.canvas);
        }
        this.canvas = null;
        this.ctx = null;
    }
}

// Create global instance
window.confetti = new Confetti();

// Convenience methods
window.fireConfetti = (options) => window.confetti.fire(options);
window.burstConfetti = () => window.confetti.burst();
window.celebrateInboxZero = () => window.confetti.celebration();
