/* MathTrail site: copy button and the live demo cards (hydrated over pre-rendered HTML). */
(function () {
  'use strict';
  Array.prototype.forEach.call(document.querySelectorAll('[data-copy]'), function (btn) {
    btn.hidden = false;
    var label = btn.querySelector('span'), idle = label.textContent;
    btn.addEventListener('click', function () {
      var done = function () { label.textContent = btn.getAttribute('data-copied'); setTimeout(function () { label.textContent = idle; }, 1600); };
      try { navigator.clipboard.writeText(btn.getAttribute('data-copy')).then(done, function () {}); } catch (e) {}
    });
  });

  var dataEl = document.getElementById('mt-data');
  if (!dataEl || !window.React || !window.ReactDOM || !window.MathTrail) return;
  var data = JSON.parse(dataEl.textContent);
  var h = React.createElement, TaskWidget = window.MathTrail.TaskWidget;
  function props(scene) {
    return { scene: scene, strings: data.strings, task: data.task, child: data.child, progress: data.progress, generatingSteps: data.gen, generatingDetail: data.genDetail };
  }
  var roots = {};
  Array.prototype.forEach.call(document.querySelectorAll('.mt-demo'), function (el) {
    roots[el.id] = ReactDOM.hydrateRoot(el, h(TaskWidget, props(el.getAttribute('data-scene'))), { identifierPrefix: el.id + '-' });
  });

  function alignFrame(el, scene) {
    if (!el) return;
    var top = 0;
    if (scene === 'wrong' || scene === 'right' || scene === 'idk') {
      var rep = el.querySelector('.mt-replies');
      if (rep) top = Math.max(0, rep.getBoundingClientRect().top - el.getBoundingClientRect().top + el.scrollTop - 180);
    } else if (scene === 'question') {
      top = el.scrollHeight;
    }
    try { el.scrollTo({ top: top, behavior: 'smooth' }); } catch (e) { el.scrollTop = top; }
  }
  setTimeout(function () {
    Array.prototype.forEach.call(document.querySelectorAll('.s-frame-scroll[data-scene]'), function (el) { alignFrame(el, el.getAttribute('data-scene')); });
  }, 300);

  // The sticky card follows the step nearest the middle of the screen.
  var sticky = document.querySelector('.s-frame-scroll[data-sticky]');
  var stickyEl = sticky && sticky.querySelector('.mt-demo');
  var stickyRoot = stickyEl && roots[stickyEl.id];
  if (!sticky || !stickyRoot) return;
  var body = sticky.querySelector('.s-frame-body'), bubble = sticky.querySelector('[data-bubble]');
  var scene = sticky.getAttribute('data-scene'), pending = null, fadeTimer = null, raf = null;
  var reduce = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  function goTo(next) {
    pending = next;
    body.classList.add('is-fading');
    clearTimeout(fadeTimer);
    fadeTimer = setTimeout(function () {
      pending = null; scene = next;
      sticky.setAttribute('data-scene', next);
      bubble.textContent = next === 'progress' ? data.bubbleProgress : data.bubble;
      stickyRoot.render(h(TaskWidget, props(next)));
      body.classList.remove('is-fading');
      setTimeout(function () { alignFrame(sticky, next); }, 80);
    }, reduce ? 0 : 160);
  }
  function onScroll() {
    if (raf) return;
    raf = requestAnimationFrame(function () {
      raf = null;
      if (sticky.offsetParent === null) return;
      var beats = document.querySelectorAll('[data-beat]'), mid = window.innerHeight / 2, best = null, bestD = Infinity;
      for (var i = 0; i < beats.length; i++) {
        var r = beats[i].getBoundingClientRect();
        if (!r.height) continue;
        var d = r.top > mid ? r.top - mid : r.bottom < mid ? mid - r.bottom : 0;
        if (d < bestD) { bestD = d; best = beats[i].getAttribute('data-beat'); }
      }
      if (best && best !== scene && best !== pending) goTo(best);
    });
  }
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll, { passive: true });
  onScroll();
})();
