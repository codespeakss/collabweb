// Title Scroller Utility
// Provides startTitleScroller, setTitleScrollerText, stopTitleScroller to animate document.title in the browser tab.

let timerId = null;
let scrollIndex = 0;
let scrollText = '';
let paused = false;
let options = {
  interval: 200, // ms between shifts
  separator: '   ', // gap between looped text
  direction: 'left' // 'left' or 'right'
};

function render() {
  if (!scrollText) return;

  const base = scrollText + options.separator;
  const n = base.length;
  if (n === 0) return;

  // Normalize index
  scrollIndex = ((scrollIndex % n) + n) % n;

  // Compute slice based on direction
  let view;
  if (options.direction === 'right') {
    // Right scrolling is equivalent to left with reversed index
    const idx = (n - scrollIndex) % n;
    view = base.slice(idx) + base.slice(0, idx);
  } else {
    view = base.slice(scrollIndex) + base.slice(0, scrollIndex);
  }

  document.title = view;
}

function tick() {
  if (paused) return;
  scrollIndex += 1;
  render();
}

function onVisibilityChange() {
  // Optionally pause when tab not visible to save CPU
  paused = document.hidden;
}

export function startTitleScroller(text, userOptions = {}) {
  stopTitleScroller();

  options = { ...options, ...userOptions };
  scrollText = (text ?? document.title ?? '').trim();
  scrollIndex = 0;
  paused = false;

  // Initial render
  render();

  // Start timer
  timerId = window.setInterval(tick, options.interval);

  // Visibility handling
  document.addEventListener('visibilitychange', onVisibilityChange);
}

export function setTitleScrollerText(text) {
  // Update the text without restarting the interval
  const newText = (text ?? '').trim();
  if (newText === scrollText) return;
  scrollText = newText;
  scrollIndex = 0;
  render();
}

export function stopTitleScroller() {
  if (timerId !== null) {
    window.clearInterval(timerId);
    timerId = null;
  }
  document.removeEventListener('visibilitychange', onVisibilityChange);
}
