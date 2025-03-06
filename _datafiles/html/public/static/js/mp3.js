class MP3Player {
  constructor(allowMultiple = false) {
    this.audioCache = new Map();
    this.currentUrl = null;
    this.loop = true;
    this.allowMultiple = allowMultiple;
    this.activeAudios = new Set();
  }

  play(url, loop = true, volume = 1.0) {
    if (!this.allowMultiple) {
      this.stopAll();
    }

    // Ensure volume doesn't exceed 1.0
    if (volume > 1.0) {
      volume = 1.0;
    }

    let audio;
    if (this.audioCache.has(url)) {
      audio = this.audioCache.get(url);
    } else {
      audio = new Audio();
      audio.src = url;
      audio.preload = "auto";
      this.audioCache.set(url, audio);
    }

    audio.loop = loop;
    audio.volume = volume;
    audio.play().catch((e) => console.error("Playback failed:", e));

    // Keep track of currently playing audio
    this.activeAudios.add(audio);

    this.currentUrl = url;
  }

  pause(url) {
    if (this.audioCache.has(url)) {
      this.audioCache.get(url).pause();
    }
  }

  stop(url) {
    if (this.audioCache.has(url)) {
      let audio = this.audioCache.get(url);
      audio.pause();
      audio.currentTime = 0;
      this.activeAudios.delete(audio);
    }
  }

  stopAll() {
    this.activeAudios.forEach(audio => {
      audio.pause();
      audio.currentTime = 0;
    });
    this.activeAudios.clear();
  }

  setVolume(url, level) {
    // Set volume for a specific URL (if needed)
    if (level < 0 || level > 1) {
      console.error("Volume must be between 0 and 1");
      return;
    }
    if (this.audioCache.has(url)) {
      this.audioCache.get(url).volume = level;
    }
  }

  setGlobalVolume(level) {
    // Set volume for any currently playing audio
    if (level < 0 || level > 1) {
      console.error("Volume must be between 0 and 1");
      return;
    }
    this.activeAudios.forEach(audio => {
      audio.volume = level;
    });
  }

  setLoop(url, loop) {
    if (this.audioCache.has(url)) {
      this.audioCache.get(url).loop = loop;
    }
  }

  isPlaying(url) {
    return this.audioCache.has(url) && !this.audioCache.get(url).paused;
  }

  getCurrentTime(url) {
    return this.audioCache.has(url) ? this.audioCache.get(url).currentTime : 0;
  }

  getDuration(url) {
    return this.audioCache.has(url) ? this.audioCache.get(url).duration : 0;
  }
}
