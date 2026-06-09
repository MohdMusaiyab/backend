const dropZone = document.getElementById('drop-zone');
const fileInput = document.getElementById('file-input');
const statusPanel = document.getElementById('status-panel');
const progressFill = document.getElementById('progress-fill');
const statusText = document.getElementById('status-text');
const results = document.getElementById('results');
const originalPreview = document.getElementById('original-preview');
const processedPreview = document.getElementById('processed-preview');

// Event Listeners for Drag & Drop
dropZone.addEventListener('click', () => fileInput.click());

['dragover', 'dragleave', 'drop'].forEach(eventName => {
  dropZone.addEventListener(eventName, preventDefaults, false);
});

function preventDefaults(e) {
  e.preventDefault();
  e.stopPropagation();
}

['dragover', 'dragenter'].forEach(eventName => {
  dropZone.addEventListener(eventName, () => dropZone.classList.add('dragover'), false);
});

['dragleave', 'drop'].forEach(eventName => {
  dropZone.addEventListener(eventName, () => dropZone.classList.remove('dragover'), false);
});

dropZone.addEventListener('drop', handleDrop, false);
fileInput.addEventListener('change', (e) => handleFiles(e.target.files), false);

function handleDrop(e) {
  const dt = e.dataTransfer;
  const files = dt.files;
  handleFiles(files);
}

function handleFiles(files) {
  if (files.length === 0) return;
  const file = files[0];
  
  if (!file.type.startsWith('image/')) {
    alert('Please upload an image file.');
    return;
  }

  // Show local preview immediately without waiting for server
  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.onloadend = () => {
    originalPreview.src = reader.result;
  };

  uploadFile(file);
}

async function uploadFile(file) {
  // Reset UI for processing
  dropZone.classList.add('hidden');
  results.classList.add('hidden');
  statusPanel.classList.remove('hidden');
  updateProgress(10, 'Uploading image...', true);

  const formData = new FormData();
  formData.append('image', file);

  try {
    const response = await fetch('/api/images/upload', {
      method: 'POST',
      body: formData
    });
    
    if (!response.ok) throw new Error('Upload failed');
    
    const data = await response.json();
    const jobId = data.jobId;
    
    // Start Polling the server for updates
    pollStatus(jobId);

  } catch (error) {
    updateProgress(0, 'Upload Failed!', false);
    console.error(error);
  }
}

async function pollStatus(jobId) {
  try {
    const response = await fetch(`/api/images/status/${jobId}`);
    const data = await response.json();

    if (data.status === 'pending') {
      updateProgress(30, 'In Queue... Waiting for an available worker', true);
      setTimeout(() => pollStatus(jobId), 1000);
    } else if (data.status === 'processing') {
      updateProgress(60, 'Worker processing image (applying filters/resizing)...', true);
      setTimeout(() => pollStatus(jobId), 1000);
    } else if (data.status === 'completed') {
      updateProgress(100, 'Processing Complete!', false);
      showResult(data.imageUrl);
    } else if (data.status === 'failed') {
      updateProgress(100, 'Worker Failed to process image', false);
      progressFill.style.background = '#ef4444'; // Red for failure
    }
  } catch (error) {
    console.error('Polling error:', error);
    // If polling fails (e.g., server restart), try again shortly
    setTimeout(() => pollStatus(jobId), 2000);
  }
}

function updateProgress(percent, text, isPulsing) {
  progressFill.style.width = `${percent}%`;
  statusText.textContent = text;
  if (isPulsing) {
    progressFill.classList.add('pulse');
  } else {
    progressFill.classList.remove('pulse');
  }
}

function showResult(imageUrl) {
  // We append a timestamp to bypass browser caching of the new image
  processedPreview.src = `${imageUrl}?t=${new Date().getTime()}`;
  results.classList.remove('hidden');
}
