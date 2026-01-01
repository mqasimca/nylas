// =============================================================================
// OTP Commands
// =============================================================================

const otpCommandSections = [
    {
        title: 'OTP Management',
        commands: {
            'get': { title: 'Get', cmd: 'otp get', desc: 'Get the latest OTP code' },
            'watch': { title: 'Watch', cmd: 'otp watch', desc: 'Watch for new OTP codes' },
            'list': { title: 'List', cmd: 'otp list', desc: 'List configured accounts' },
            'messages': { title: 'Messages', cmd: 'otp messages', desc: 'Show recent OTP messages' }
        }
    }
];

const otpCommands = {};
otpCommandSections.forEach(section => {
    Object.assign(otpCommands, section.commands);
});

let currentOtpCmd = '';

function showOtpCmd(cmd) {
    const data = otpCommands[cmd];
    if (!data) return;

    currentOtpCmd = cmd;

    document.querySelectorAll('#page-otp .cmd-item').forEach(item => {
        item.classList.toggle('active', item.dataset.cmd === cmd);
    });

    const detail = document.getElementById('otp-detail');
    detail.querySelector('.detail-placeholder').style.display = 'none';
    detail.querySelector('.detail-content').style.display = 'block';

    document.getElementById('otp-detail-title').textContent = data.title;
    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + data.cmd;
    document.getElementById('otp-detail-desc').textContent = data.desc || '';
    document.getElementById('otp-output').textContent = 'Click "Run" to execute command...';
    document.getElementById('otp-output').className = 'output-pre';

    showParamInput('otp', data.param, data.flags);
}

async function runOtpCmd() {
    if (!currentOtpCmd) return;

    const data = otpCommands[currentOtpCmd];
    const output = document.getElementById('otp-output');
    const btn = document.getElementById('otp-run-btn');
    const fullCmd = buildCommand(data.cmd, 'otp', data.flags);

    document.getElementById('otp-detail-cmd').textContent = 'nylas ' + fullCmd;

    setButtonLoading(btn, true);
    setOutputLoading(output);

    try {
        const res = await fetch('/api/exec', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ command: fullCmd })
        });
        const result = await res.json();

        if (result.error) {
            setOutputError(output, result.error);
            showToast('Command failed', 'error');
        } else {
            setOutputSuccess(output, result.output);
            showToast('Command completed', 'success');
        }

        updateTimestamp('otp');
    } catch (err) {
        setOutputError(output, 'Failed to execute command: ' + err.message);
        showToast('Connection error', 'error');
    } finally {
        setButtonLoading(btn, false);
    }
}

function refreshOtpCmd() {
    if (currentOtpCmd) runOtpCmd();
}

function renderOtpCommands() {
    renderCommandSections('otp-cmd-list', otpCommandSections, 'showOtpCmd');
}
