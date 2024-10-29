<template>
  <div class="wrap">
    <div class="tri" v-for="n in 64" :key="n"></div>

    <!-- Logo Image -->
    <div id="logo">
      <img src="/assets/logo_white_outline.svg" alt="Blazium Engine Logo" />
    </div>

    <!-- Intro Section -->
    <div id="intro">
      <h2 class="intro-title">Blazium Engine</h2>
      <hr />
      <div class="social-icons">
        <a 
          title="Discord / Community" 
          href="https://discord.gg/Jk93kePyty" 
          target="_blank" 
          rel="noopener noreferrer"
        >
          <i class="fab fa-discord"></i>
        </a>
        <a 
          title="Source Code" 
          href="https://github.com/blazium-engine/blazium" 
          target="_blank" 
          rel="noopener noreferrer"
        >
          <i class="fab fa-github"></i>
        </a>
        <a 
          title="Twitter / X" 
          href="https://x.com/BlaziumEngine" 
          target="_blank" 
          rel="noopener noreferrer"
        >
          <i class="fab fa-twitter"></i>
        </a>
        <a 
          title="Documentation" 
          href="https://docs.blazium.app" 
          target="_blank" 
          rel="noopener noreferrer"
        >
          <i class="fas fa-book"></i>
        </a>
      </div>
    </div>

    <!-- Release Information Box -->
    <div id="release-info">
      <h3 class="release-title">Blazium Engine Release Info</h3>
      <p id="release-message">{{ releaseMessage }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import "@/assets/css/style.css";
import "@/assets/css/effects.css";
import "@fortawesome/fontawesome-free/css/all.min.css";

// Dynamically import background.js on client-side
if (typeof window !== 'undefined') {
  import("@/assets/js/background.js");
}

// Reactive state for release message
const releaseMessage = ref("Loading release information...");


// Define meta tags using useHead (Assuming you're using Nuxt 3 or vueuse/head)
import { useHead } from '@vueuse/head';

useHead({
  title: "Blazium Engine",
  meta: [
    { charset: "utf-8" },
    { name: "viewport", content: "width=device-width, initial-scale=1.0" },
    { 
      name: "description", 
      content: "Blazium Engine - A powerful and flexible game engine for developers." 
    },
    { 
      name: "keywords", 
      content: "Blazium, Game Engine, Game Development, Blazium Engine, Blazium App" 
    },
    { name: "robots", content: "index, follow" },
    { name: "author", content: "Blazium Team" },
    { name: "theme-color", content: "#000000" },
    { property: "og:title", content: "Blazium Engine" },
    { 
      property: "og:description", 
      content: "Blazium Engine - A powerful and flexible game engine for developers." 
    },
    { property: "og:url", content: "https://blazium.app/" },
    { property: "og:type", content: "website" },
    { property: "og:site_name", content: "Blazium Engine" },
  ],
  link: [
    { rel: "shortcut icon", type: "image/svg+xml", href: "/assets/logo.svg" },
    { 
      href: "https://fonts.googleapis.com/css2?family=Montserrat:wght@400;700&display=swap",
      rel: "stylesheet" 
    }
  ]
});
</script>

<style>
/* CSS Variables for easy theming */
:root {
  --background-color: #1e1e1e;
  --text-color: #f0f0f0;
  --accent-color: #e83951;
  --font-family: 'Montserrat', sans-serif;
}

/* Wrapper styling */
.wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--background-color);
  color: var(--text-color);
  text-align: center;
  padding: 2rem;
  position: relative;
  overflow: hidden;
}

/* Logo styling with hover animation */
#logo img {
  width: 280px;
  transition: transform 0.5s ease;
}

#logo img:hover {
  transform: rotate(20deg) scale(1.1);
}

/* Intro title styling */
.intro-title {
  font-weight: bold;
  margin: 1em 0;
}

/* Social icons container */
.social-icons {
  display: flex;
  justify-content: center;
  margin: 1em 0;
}

/* Social icons styling */
.social-icons a {
  font-size: 2.0em;
  color: var(--accent-color);
  margin: 0 1rem;
  transition: color 0.3s, transform 0.3s;
}

.social-icons a:hover {
  color: #ffffff;
  transform: translateY(-5px);
}

/* Release info section */
#release-info {
  margin-top: 20px;
}

/* Release title styling */
.release-title {
  font-family: var(--font-family);
  font-weight: bold;
  font-size: 1.6em;
  margin-bottom: 0.5em;
}

/* Release message styling */
.release-message {
  font-family: var(--font-family);
  font-size: 1.4em;
  margin-bottom: 0.5em;
}


/* Responsive Design */
@media (max-width: 600px) {
  .intro-title {
    font-size: 2.4em;
  }

  .social-icons a {
    font-size: 2.0em;
    margin: 0 0.5rem;
  }

  #logo img {
    width: 100px;
  }

  #release-info {
    width: 80%;
  }
}
</style>