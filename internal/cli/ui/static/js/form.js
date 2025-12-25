// =============================================================================
// Setup Form
// =============================================================================

function initForm() {
    const form = document.getElementById('setup-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const btn = form.querySelector('.btn-primary');
        const error = document.getElementById('error-msg');

        error?.classList.remove('visible');
        btn.classList.add('loading');
        btn.disabled = true;

        try {
            const res = await fetch('/api/config/setup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    api_key: document.getElementById('api-key').value.trim(),
                    region: document.getElementById('region').value
                })
            });
            const data = await res.json();

            if (data.success) {
                showDashboard(data);
            } else {
                showFormError(data.error || 'Setup failed');
            }
        } catch (err) {
            showFormError('Connection failed. Please try again.');
        } finally {
            btn.classList.remove('loading');
            btn.disabled = false;
        }
    });
}

function showFormError(msg) {
    const error = document.getElementById('error-msg');
    if (error) {
        error.textContent = msg;
        error.classList.add('visible');
    }
}

function togglePassword() {
    const input = document.getElementById('api-key');
    const btn = input.parentElement.querySelector('.input-btn');
    const isPassword = input.type === 'password';
    input.type = isPassword ? 'text' : 'password';
    btn?.classList.toggle('visible', isPassword);
}
