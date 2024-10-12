document.addEventListener('DOMContentLoaded', function () {
  particlesJS("particles", {
    "particles":{
    "number":{
      "value":256,"density":{
        "enable":true,"value_area":974
      }
    },
    "color":{
      "value":"#ffffff"
    },
    "shape":{
      "type":"circle","stroke":{
        "width":0,"color":"#000000"
      },
      "polygon":{
        "nb_sides":5
      },
      "image":{
        "src":"img/github.svg","width":100,"height":100
      }
    },
    "opacity":{
      "value":0.5,"random":false,"anim":{
        "enable":false,"speed":1,"opacity_min":0.1,"sync":false
      }
    },
    "size":{
      "value":3,"random":true,"anim":{
        "enable":false,"speed":40,"size_min":0.1,"sync":false
      }
    },
    "line_linked":{
      "enable":true,"distance":150,"color":"#ffffff","opacity":0.4,"width":1
    },
    "move":{
      "enable":true,"speed":6,"direction":"none","random":false,"straight":false,"out_mode":"out","bounce":false,
      "attract":{
        "enable":false,"rotateX":600,"rotateY":1200
      }
    }
  },
  "interactivity":{
    "detect_on":"canvas","events":{
      "onhover":{
        "enable":true,"mode":"repulse"
      },
      "onclick":{
        "enable":true,"mode":"push"
      },
      "resize":true
    },"modes":{
      "grab":{
        "distance":400,"line_linked":{
          "opacity":1
        }
      },
      "bubble":{
        "distance":400,"size":40,"duration":2,"opacity":8,"speed":3
      },
      "repulse":{
        "distance":200,"duration":0.4
      },
      "push":{
        "particles_nb":4
      },"remove":{
        "particles_nb":2
      }
    }
  },"retina_detect":true
});

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
        releaseMessage.innerHTML = 'Release is pending, visit our <a href="https://discord.gg/sZaf9KYzDp" style="color: #E83951;">Discord</a> for more information.';
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

  var intro = document.getElementById('intro');
  intro.style.marginTop = - intro.offsetHeight / 2 + 'px';
}, false);
