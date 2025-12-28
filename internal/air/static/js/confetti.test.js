/**
 * Confetti Unit Tests
 * Run in browser console: copy-paste this file or load as a script
 * Usage: ConfettiTests.run()
 */

const ConfettiTests = {
    passed: 0,
    failed: 0,
    results: [],

    assert(condition, message) {
        if (condition) {
            this.passed++;
            this.results.push({ status: 'PASS', message });
            console.log(`✓ ${message}`);
        } else {
            this.failed++;
            this.results.push({ status: 'FAIL', message });
            console.error(`✗ ${message}`);
        }
    },

    assertEquals(actual, expected, message) {
        this.assert(actual === expected, `${message} (expected: ${expected}, got: ${actual})`);
    },

    assertExists(value, message) {
        this.assert(value !== undefined && value !== null, message);
    },

    // Test: Confetti class exists
    testConfettiClassExists() {
        this.assertExists(window.Confetti, 'Confetti class should exist on window');
    },

    // Test: Global instance exists
    testGlobalInstanceExists() {
        this.assertExists(window.confetti, 'Global confetti instance should exist');
        this.assert(window.confetti instanceof Confetti, 'confetti should be instance of Confetti');
    },

    // Test: Convenience methods exist
    testConvenienceMethodsExist() {
        this.assertExists(window.fireConfetti, 'fireConfetti should exist');
        this.assertExists(window.burstConfetti, 'burstConfetti should exist');
        this.assertExists(window.celebrateInboxZero, 'celebrateInboxZero should exist');
    },

    // Test: Constructor sets default options
    testDefaultOptions() {
        const confetti = new Confetti();
        this.assertEquals(confetti.options.particleCount, 100, 'Default particleCount should be 100');
        this.assertEquals(confetti.options.spread, 70, 'Default spread should be 70');
        this.assertEquals(confetti.options.decay, 0.95, 'Default decay should be 0.95');
        this.assertEquals(confetti.options.gravity, 1, 'Default gravity should be 1');
        this.assertEquals(confetti.options.ticks, 200, 'Default ticks should be 200');
        this.assert(Array.isArray(confetti.options.colors), 'Colors should be an array');
        this.assert(confetti.options.colors.length > 0, 'Colors should have items');
    },

    // Test: Constructor accepts custom options
    testCustomOptions() {
        const confetti = new Confetti({
            particleCount: 50,
            spread: 100,
            gravity: 2
        });
        this.assertEquals(confetti.options.particleCount, 50, 'Custom particleCount should be set');
        this.assertEquals(confetti.options.spread, 100, 'Custom spread should be set');
        this.assertEquals(confetti.options.gravity, 2, 'Custom gravity should be set');
        // Defaults should still apply for unset options
        this.assertEquals(confetti.options.decay, 0.95, 'Default decay should apply');
    },

    // Test: Initial state
    testInitialState() {
        const confetti = new Confetti();
        this.assert(confetti.canvas === null, 'Canvas should initially be null');
        this.assert(confetti.ctx === null, 'Context should initially be null');
        this.assert(Array.isArray(confetti.particles), 'Particles should be an array');
        this.assertEquals(confetti.particles.length, 0, 'Particles should be empty initially');
        this.assert(confetti.isRunning === false, 'Should not be running initially');
        this.assert(confetti.animationId === null, 'Animation ID should be null initially');
    },

    // Test: createParticle returns valid particle
    testCreateParticle() {
        const confetti = new Confetti();
        const particle = confetti.createParticle(100, 100);

        this.assertExists(particle, 'Particle should be created');
        this.assertEquals(particle.x, 100, 'Particle x should match');
        this.assertEquals(particle.y, 100, 'Particle y should match');
        this.assertExists(particle.vx, 'Particle should have vx');
        this.assertExists(particle.vy, 'Particle should have vy');
        this.assertExists(particle.color, 'Particle should have color');
        this.assertExists(particle.shape, 'Particle should have shape');
        this.assertExists(particle.size, 'Particle should have size');
        this.assert(particle.size > 0, 'Particle size should be positive');
        this.assertExists(particle.rotation, 'Particle should have rotation');
        this.assertExists(particle.rotationSpeed, 'Particle should have rotationSpeed');
        this.assertEquals(particle.ticks, 200, 'Particle ticks should match options');
        this.assertEquals(particle.opacity, 1, 'Particle opacity should start at 1');
    },

    // Test: updateParticle modifies particle correctly
    testUpdateParticle() {
        const confetti = new Confetti();
        const particle = confetti.createParticle(100, 100);
        const initialTicks = particle.ticks;
        const initialOpacity = particle.opacity;

        const alive = confetti.updateParticle(particle);

        this.assert(alive === true, 'Particle should be alive after first update');
        this.assertEquals(particle.ticks, initialTicks - 1, 'Ticks should decrease by 1');
        this.assert(particle.opacity < initialOpacity, 'Opacity should decrease');
    },

    // Test: updateParticle returns false when ticks reach 0
    testUpdateParticleExpires() {
        const confetti = new Confetti();
        const particle = confetti.createParticle(100, 100);
        particle.ticks = 1;

        const alive = confetti.updateParticle(particle);

        this.assertEquals(particle.ticks, 0, 'Ticks should be 0');
        this.assert(alive === false, 'Particle should be dead when ticks reach 0');
    },

    // Test: fire() creates canvas and particles
    testFireCreatesParticles() {
        const confetti = new Confetti();
        confetti.fire({ particleCount: 10 });

        this.assertExists(confetti.canvas, 'Canvas should be created');
        this.assertExists(confetti.ctx, 'Context should be created');
        this.assert(confetti.particles.length > 0, 'Particles should be created');
        this.assert(confetti.isRunning, 'Animation should be running');

        // Clean up
        confetti.destroy();
    },

    // Test: stop() clears particles and state
    testStopClearsState() {
        const confetti = new Confetti();
        confetti.fire({ particleCount: 10 });
        confetti.stop();

        this.assertEquals(confetti.particles.length, 0, 'Particles should be cleared');
        this.assert(confetti.isRunning === false, 'Should not be running after stop');
        this.assert(confetti.animationId === null, 'Animation ID should be null after stop');

        // Clean up
        confetti.destroy();
    },

    // Test: destroy() removes canvas from DOM
    testDestroyRemovesCanvas() {
        const confetti = new Confetti();
        confetti.fire({ particleCount: 5 });

        this.assertExists(confetti.canvas, 'Canvas should exist before destroy');
        const canvasInDom = document.body.contains(confetti.canvas);
        this.assert(canvasInDom, 'Canvas should be in DOM before destroy');

        confetti.destroy();

        this.assert(confetti.canvas === null, 'Canvas should be null after destroy');
        this.assert(confetti.ctx === null, 'Context should be null after destroy');
    },

    // Test: resizeCanvas updates dimensions
    testResizeCanvas() {
        const confetti = new Confetti();
        confetti.fire({ particleCount: 1 });

        confetti.resizeCanvas();

        this.assertEquals(confetti.canvas.width, window.innerWidth, 'Canvas width should match window');
        this.assertEquals(confetti.canvas.height, window.innerHeight, 'Canvas height should match window');

        // Clean up
        confetti.destroy();
    },

    // Test: burst() uses center origin
    testBurstUsesCenter() {
        const confetti = new Confetti();
        // Override fire to capture options
        const originalFire = confetti.fire.bind(confetti);
        let capturedOptions = null;
        confetti.fire = (opts) => {
            capturedOptions = opts;
            // Don't actually fire to avoid animation
        };

        confetti.burst();

        this.assertExists(capturedOptions, 'Options should be captured');
        this.assertEquals(capturedOptions.origin.x, 0.5, 'Burst origin.x should be 0.5');
        this.assertEquals(capturedOptions.origin.y, 0.5, 'Burst origin.y should be 0.5');
        this.assertEquals(capturedOptions.spread, 360, 'Burst spread should be 360');

        // Restore original
        confetti.fire = originalFire;
    },

    // Test: color is from options array
    testParticleColorFromOptions() {
        const customColors = ['#ff0000', '#00ff00', '#0000ff'];
        const confetti = new Confetti({ colors: customColors });
        const particle = confetti.createParticle(100, 100);

        this.assert(customColors.includes(particle.color), 'Particle color should be from options');
    },

    // Test: shape is from options array
    testParticleShapeFromOptions() {
        const customShapes = ['triangle', 'star'];
        const confetti = new Confetti({ shapes: customShapes });
        const particle = confetti.createParticle(100, 100);

        this.assert(customShapes.includes(particle.shape), 'Particle shape should be from options');
    },

    // Run all tests
    run() {
        console.log('🎉 Running Confetti Unit Tests...\n');

        this.passed = 0;
        this.failed = 0;
        this.results = [];

        // Run all test methods
        const testMethods = Object.getOwnPropertyNames(ConfettiTests)
            .filter(name => name.startsWith('test'));

        for (const method of testMethods) {
            try {
                console.log(`\n--- ${method} ---`);
                this[method]();
            } catch (error) {
                this.failed++;
                this.results.push({ status: 'ERROR', message: `${method}: ${error.message}` });
                console.error(`✗ ${method}: ${error.message}`);
            }
        }

        // Summary
        console.log('\n========================================');
        console.log(`Tests completed: ${this.passed + this.failed}`);
        console.log(`Passed: ${this.passed}`);
        console.log(`Failed: ${this.failed}`);
        console.log('========================================\n');

        if (this.failed === 0) {
            console.log('🎉 All tests passed!');
        } else {
            console.log('❌ Some tests failed');
        }

        return { passed: this.passed, failed: this.failed, results: this.results };
    }
};

// Export for use
if (typeof window !== 'undefined') {
    window.ConfettiTests = ConfettiTests;
}

// Auto-run if loaded as script with ?run parameter
if (typeof location !== 'undefined' && location.search.includes('run')) {
    document.addEventListener('DOMContentLoaded', () => ConfettiTests.run());
}
