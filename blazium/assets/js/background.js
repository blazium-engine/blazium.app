function fetchReleaseInfo() {
  fetch('/mirror/mirrors.json')
    .then(response => response.json())
    .then(data => {
      const releaseMessage = document.getElementById('release-message');

      if (data.mirrors.length === 0) {
        releaseMessage.innerHTML = 'Release is pending, visit our <a href="https://chat.blazium.app" style="color: #E83951;">Discord</a> for more information.';
      } else {
        const mirrors = data.mirrors.map(mirror => `
          <li>
            <a href="${mirror.download_url}" style="color: #E83951;">Download</a> |
            SHA256: ${mirror.sha} |
            Release Date: ${mirror.release_date}
          </li>
        `).join('');

        releaseMessage.innerHTML = `
          <ul style="list-style: none; padding-left: 0; text-align: center;">
            ${mirrors}
          </ul>
        `;
      }
    })
    .catch(error => {
      console.error('Error fetching release information:', error);
      document.getElementById('release-message').innerHTML = 'Error fetching release information. Please try again later.';
    });
}

// Call the function to fetch the release info when the page loads
window.onload = fetchReleaseInfo;
