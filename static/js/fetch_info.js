function fetchReleaseInfo() {
  const protocol = window.location.protocol;
  const host = window.location.host;
  const apiUrl = `${protocol}//${host}/api/mirrorlist/latest/json`;

  console.log("Fetching from URL:", apiUrl);  // Add this line to check the URL

  fetch(apiUrl)
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok ' + response.statusText);
      }
      return response.json();
    })
    .then(data => {
      const releaseMessage = document.getElementById('release-message');
      
      if (data.mirrors.length === 0) {
        releaseMessage.innerHTML = 'Release is pending, visit our <a href="https://chat.blazium.app">Discord</a> for more information.';
      } else {
        const mirrors = data.mirrors.map(mirror => `
          <li>
            <a href="${mirror.download_url}">Download</a> |
            SHA256: ${mirror.sha} |
            Release Date: ${mirror.release_date}
          </li>
        `).join('');
        
        releaseMessage.innerHTML = `
          <ul>
            ${mirrors}
          </ul>
        `;
      }
    })
    .catch(error => {
      console.error('Error fetching release information:', error);
      document.getElementById('release-message').innerHTML = 'Error fetching release information. Please try again later.';
    });
};

// Call the function to fetch the release info when the page loads
window.onload = fetchReleaseInfo;
