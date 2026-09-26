/* @ds-bundle: {"format":4,"namespace":"MathTrail","components":[{"name":"TaskWidget"},{"name":"OptionRow"},{"name":"OptionList"},{"name":"Button"},{"name":"ReplyField"},{"name":"ThreadBar"},{"name":"MessageHeader"},{"name":"Badge"},{"name":"Diagram"},{"name":"Note"},{"name":"Verdict"},{"name":"SolutionSteps"},{"name":"ReplyCard"},{"name":"GeneratingSteps"},{"name":"RatingSummary"},{"name":"StatList"},{"name":"ProfileFields"},{"name":"Icon"},{"name":"Mark"},{"name":"Avatar"}]} */
(function () {
  'use strict';

  function R() { return window.React; }
  function h() { var r = R(); return r.createElement.apply(r, arguments); }
  function cx() { return Array.prototype.filter.call(arguments, Boolean).join(' '); }
  function fmt(s, vars) {
    return String(s == null ? '' : s).replace(/\{(\w+)\}/g, function (m, k) { return vars && vars[k] != null ? vars[k] : m; });
  }

  /* ---------- Icons: inline stroke SVG, colour from currentColor ---------- */
  var S16 = { fill: 'none', stroke: 'currentColor', strokeLinecap: 'round', strokeLinejoin: 'round' };
  function p16(d, w) { return h('path', Object.assign({ d: d, strokeWidth: w }, S16)); }
  var ICONS = {
    'chevron-right': function () { return p16('M6 4l4 4-4 4', 1.75); },
    'chevron-left': function () { return p16('M10 4l-4 4 4 4', 1.75); },
    'send': function () { return p16('M3 8h9.5M8.5 4l4 4-4 4', 1.75); },
    'check': function () { return p16('M3.5 8.5l3 3 6-7', 2); },
    'cross': function () { return p16('M4.5 4.5l7 7M11.5 4.5l-7 7', 2); },
    'hint': function () { return p16('M6 12.5h4M6.5 14.5h3M8 1.5a4.5 4.5 0 0 0-2.6 8.2c.4.3.6.7.6 1.1v.2h4v-.2c0-.4.2-.8.6-1.1A4.5 4.5 0 0 0 8 1.5z', 1.4); },
    'trap': function () { return p16('M8 2.2L14.3 13.3H1.7L8 2.2zM8 6.5v3M8 11.4v.1', 1.4); },
    'more': function () {
      return h(R().Fragment, null,
        h('circle', { cx: 3, cy: 8, r: 1.5, fill: 'currentColor' }),
        h('circle', { cx: 8, cy: 8, r: 1.5, fill: 'currentColor' }),
        h('circle', { cx: 13, cy: 8, r: 1.5, fill: 'currentColor' }));
    }
  };

  function Icon(props) {
    var name = props.name, size = props.size || 16;
    if (name === 'spinner') {
      return h('svg', { width: size, height: size, viewBox: '0 0 20 20', fill: 'none', 'aria-hidden': 'true', className: cx('mt-icon mt-spin', props.className) },
        h('circle', { cx: 10, cy: 10, r: 8, stroke: props.track ? 'var(--' + props.track + ')' : 'var(--track)', strokeWidth: 2.5 }),
        h('path', { d: 'M10 2a8 8 0 0 1 8 8', stroke: 'currentColor', strokeWidth: 2.5, strokeLinecap: 'round' }));
    }
    if (name === 'step-done' || name === 'verdict-correct' || name === 'verdict-wrong') {
      var fill = name === 'step-done' ? 'var(--ink-strong)' : name === 'verdict-correct' ? 'var(--correct)' : 'var(--wrong)';
      var ink = name === 'step-done' ? 'var(--surface)' : 'var(--on-signal)';
      var mark = name === 'verdict-wrong' ? 'M7 7l6 6M13 7l-6 6' : 'M6 10.2l2.6 2.6L14 7.4';
      return h('svg', { width: size, height: size, viewBox: '0 0 20 20', fill: 'none', 'aria-hidden': 'true', className: cx('mt-icon', props.className) },
        h('circle', { cx: 10, cy: 10, r: 10, fill: fill }),
        h('path', { d: mark, stroke: ink, strokeWidth: 2, strokeLinecap: 'round', strokeLinejoin: 'round' }));
    }
    if (name === 'step-waiting') {
      return h('svg', { width: size, height: size, viewBox: '0 0 20 20', fill: 'none', 'aria-hidden': 'true', className: cx('mt-icon', props.className) },
        h('circle', { cx: 10, cy: 10, r: 8.5, stroke: 'var(--track)', strokeWidth: 1.5 }));
    }
    var draw = ICONS[name];
    if (!draw) return null;
    return h('svg', { width: size, height: size, viewBox: '0 0 16 16', fill: 'none', 'aria-hidden': 'true', className: cx('mt-icon', props.className) }, draw());
  }

  /* ---------- Brand ---------- */
  function Mark(props) {
    var size = props.size || 32;
    return h('svg', { width: size, height: size, viewBox: '0 0 32 32', 'aria-hidden': props.label ? undefined : 'true', role: props.label ? 'img' : undefined, 'aria-label': props.label, className: 'mt-icon' },
      h('circle', { cx: 16, cy: 16, r: 16, fill: 'var(--mark-fill)' }),
      h('path', { d: 'M8.5 21.5L13.5 16.5L17 19L22.5 12.5', fill: 'none', stroke: 'var(--mark-ink)', strokeWidth: 2.2, strokeLinecap: 'round', strokeLinejoin: 'round' }),
      h('circle', { cx: 23, cy: 12, r: 2.6, fill: 'var(--mark-ink)' }));
  }

  function Avatar(props) {
    var size = props.size || 24;
    return h('svg', { width: size, height: size, viewBox: '0 0 24 24', 'aria-hidden': 'true', className: 'mt-icon' },
      h('circle', { cx: 12, cy: 12, r: 12, fill: 'var(--avatar-fill)' }),
      h('circle', { cx: 12, cy: 9.5, r: 3.6, fill: 'var(--avatar-ink)' }),
      h('path', { d: 'M5 19.6c1.3-3.2 3.9-4.8 7-4.8s5.7 1.6 7 4.8', fill: 'var(--avatar-ink)' }));
  }

  /* ---------- Controls ---------- */
  function Button(props) {
    var variant = props.variant || 'secondary';
    return h('button', {
      type: 'button',
      className: cx('mt-btn', variant === 'primary' && 'mt-btn-primary', props.className),
      disabled: !!props.disabled,
      'aria-expanded': props.pressed == null ? undefined : String(!!props.pressed),
      onClick: props.onClick
    }, props.children);
  }

  var STATUS_ICON = { selected: 'spinner', correct: 'check', wrong: 'cross' };
  function OptionRow(props) {
    var state = props.state || 'default';
    var locked = state === 'correct' || state === 'wrong' || state === 'muted' || props.disabled;
    var icon = STATUS_ICON[state];
    return h('label', { className: 'mt-option', 'data-state': state },
      h('input', {
        type: 'radio', name: props.name || 'answer', value: props.letter, className: 'mt-vh',
        checked: state === 'selected' || (state === 'wrong') || !!props.checked,
        disabled: locked,
        onChange: function () { if (!locked && props.onSelect) props.onSelect(props.letter); }
      }),
      h('span', { className: 'mt-option-letter' }, props.letter),
      h('span', { className: 'mt-option-value' }, props.value),
      props.status ? h('span', { className: 'mt-option-status' },
        icon ? h(Icon, { name: icon, size: 16, track: state === 'selected' ? 'accent-tint' : undefined }) : null,
        props.status) : null);
  }

  function OptionList(props) {
    return h('fieldset', { className: 'mt-options', disabled: !!props.disabled },
      h('legend', null, props.legend),
      h('div', { className: 'mt-options-list' }, (props.options || []).map(function (o) {
        return h(OptionRow, Object.assign({ key: o.letter, name: props.name, onSelect: props.onSelect }, o));
      })));
  }

  function ReplyField(props) {
    var st = R().useState('');
    var value = props.value != null ? props.value : st[0];
    var setValue = function (v) { if (props.onChange) props.onChange(v); st[1](v); };
    var id = R().useId ? R().useId() : 'mt-ask';
    var send = function () {
      if (props.disabled || !String(value).trim()) return;
      if (props.onSend) props.onSend(String(value).trim());
      setValue('');
    };
    return h('div', { className: 'mt-field', 'data-disabled': props.disabled ? 'true' : undefined },
      h('label', { htmlFor: id, className: 'mt-vh' }, props.label || props.placeholder),
      h('input', {
        id: id, type: 'text', placeholder: props.placeholder, disabled: !!props.disabled, value: value,
        onChange: function (e) { setValue(e.target.value); },
        onKeyDown: function (e) { if (e.key === 'Enter') { e.preventDefault(); send(); } }
      }),
      h('button', { type: 'button', className: 'mt-icon-btn', 'aria-label': props.sendLabel || 'Send', disabled: props.disabled || !String(value).trim(), onClick: send },
        h(Icon, { name: 'send', size: 18 })));
  }

  /* ---------- Thread ---------- */
  function ThreadBar(props) {
    if (props.variant === 'back') {
      return h('button', { type: 'button', className: 'mt-bar mt-bar-back', onClick: props.onClick },
        h(Icon, { name: 'chevron-left', size: 16, className: 'mt-chevron' }),
        h('span', null, props.label));
    }
    return h('button', { type: 'button', className: 'mt-bar', onClick: props.onClick },
      h(Avatar, { size: 24 }),
      h('span', { className: 'mt-bar-name' }, props.name),
      h('span', { className: 'mt-bar-action' }, props.action),
      h(Icon, { name: 'chevron-right', size: 16, className: 'mt-chevron' }));
  }

  function Badge(props) { return h('span', { className: 'mt-badge' }, props.children); }

  function MessageHeader(props) {
    var author = props.author || 'app';
    var compact = !!props.compact;
    var pic = author === 'app' ? h(Mark, { size: compact ? 28 : 32 }) : h(Avatar, { size: compact ? 24 : 32 });
    var nameLine = [h('span', { key: 'n', className: 'mt-name' }, props.name)];
    if (props.time) nameLine.push(h('span', { key: 't', className: 'mt-meta' }, props.time));
    var text;
    if (compact) {
      text = h('div', { className: 'mt-head-text' }, nameLine);
    } else if (props.wide) {
      text = h('div', { className: 'mt-head-text' },
        h('span', { className: 'mt-name' }, props.name),
        props.badge ? h(Badge, null, props.badge) : null,
        props.time ? h('span', { className: 'mt-meta' }, props.time) : null);
    } else {
      text = h('div', { className: 'mt-head-text' },
        h('div', { className: 'mt-head-line' }, nameLine),
        props.badge ? h(Badge, null, props.badge) : null);
    }
    return h('header', { className: cx('mt-head', compact && 'mt-head-compact', props.wide && 'mt-head-wide') },
      pic, text,
      h('button', { type: 'button', className: 'mt-icon-btn mt-icon-btn-44', 'aria-label': props.moreLabel || 'More options', 'aria-haspopup': 'menu' },
        h(Icon, { name: 'more', size: 16 })));
  }

  function ReplyCard(props) {
    return h('article', { className: 'mt-reply' },
      h(MessageHeader, { author: props.author, name: props.name, time: props.time, compact: true, moreLabel: props.moreLabel }),
      h('div', { className: cx('mt-reply-body', props.tight && 'mt-reply-body-tight') }, props.children));
  }

  /* ---------- Task blocks ---------- */
  function Diagram(props) { return h('pre', { dir: 'ltr', className: 'mt-diagram' }, props.children); }

  function Note(props) {
    var tone = props.tone || 'plain';
    var icon = tone === 'hint' ? 'hint' : tone === 'trap' ? 'trap' : null;
    return h(tone === 'hint' ? 'aside' : 'div', { className: cx('mt-note', 'mt-note-' + tone) },
      h('div', { className: 'mt-note-label' }, icon ? h(Icon, { name: icon, size: 16 }) : null, h('span', null, props.label)),
      h('p', null, props.children));
  }

  function Verdict(props) {
    var tone = props.tone || 'none';
    var icon = tone === 'correct' ? 'verdict-correct' : tone === 'wrong' ? 'verdict-wrong' : null;
    var line = h('p', { className: 'mt-verdict-line' }, icon ? h(Icon, { name: icon, size: 20 }) : null, h('span', null, props.children));
    if (!props.detail) return line;
    return h('div', { className: 'mt-verdict' }, line, h('p', { className: 'mt-verdict-detail' }, props.detail));
  }

  function SolutionSteps(props) {
    return h('div', { className: 'mt-steps' },
      h('h3', { className: 'mt-section-label' }, props.label),
      h('ol', null, (props.steps || []).map(function (s, i) {
        return h('li', { key: i }, h('span', { className: 'mt-step-num' }, i + 1), h('span', null, s));
      })));
  }

  function GeneratingSteps(props) {
    var labels = props.statusLabels || { done: 'Done:', active: 'In progress:', waiting: 'Waiting:' };
    return h('div', { className: 'mt-gen' },
      h('p', { className: 'mt-gen-title' }, props.title),
      h('ol', { 'aria-live': 'polite' }, (props.steps || []).map(function (s, i) {
        var status = s.status || 'waiting';
        var icon = status === 'done' ? h(Icon, { name: 'step-done', size: 20 })
          : status === 'active' ? h('span', { style: { color: 'var(--ink-strong)', display: 'flex' } }, h(Icon, { name: 'spinner', size: 20 }))
          : h(Icon, { name: 'step-waiting', size: 20 });
        return h('li', { key: i, 'data-status': status },
          h('span', { className: 'mt-gen-icon' }, icon),
          h('span', { className: 'mt-gen-label' },
            h('span', { className: 'mt-vh' }, labels[status] + ' '),
            s.label,
            s.detail && status === 'active' ? h('span', { className: 'mt-gen-detail' }, ' ' + s.detail) : null));
      })));
  }

  /* ---------- Progress ---------- */
  function RatingSummary(props) {
    var total = props.total || 5, rank = props.rank || 0, pips = [];
    for (var i = 0; i < total; i++) pips.push(h('span', { key: i, className: 'mt-pip', 'data-on': String(i < rank) }));
    return h('section', { className: 'mt-rating', 'aria-label': props.ariaLabel || 'Overall rating' },
      h('span', { className: 'mt-rating-rank' }, props.rankLabel),
      h('div', { className: 'mt-rating-line' },
        h('span', { className: 'mt-rating-num' }, props.rating),
        h('span', { className: 'mt-meta' }, props.ratingLabel)),
      h('div', { className: 'mt-pips', 'aria-hidden': 'true' }, pips));
  }

  function StatusMark(props) {
    var tone = props.tone;
    return h('span', { className: cx('mt-status', 'mt-status-' + tone) }, h(Icon, { name: tone === 'correct' ? 'check' : 'cross', size: 16 }), props.label);
  }

  function StatList(props) {
    return h('section', { className: 'mt-list' },
      h('h3', { className: 'mt-section-label' }, props.label),
      h('ul', null, (props.rows || []).map(function (r, i) {
        if (r.bar != null) {
          return h('li', { key: i, className: 'mt-row mt-row-bar' },
            h('div', { className: 'mt-row-bar-top' },
              h('span', { className: 'mt-row-label' }, r.label),
              h('span', { className: 'mt-row-spacer' }),
              r.count ? h('span', { className: 'mt-row-count' }, r.count) : null),
            h('div', { className: 'mt-bar-track', 'aria-hidden': 'true' }, h('div', { className: 'mt-bar-fill', style: { width: Math.max(0, Math.min(100, r.bar)) + '%' } })));
        }
        return h('li', { key: i, className: 'mt-row' },
          h('span', { className: 'mt-row-label' }, r.label),
          h('span', { className: 'mt-row-spacer' }),
          r.status ? h(StatusMark, r.status) : null,
          r.value != null ? h('span', { className: 'mt-row-value' }, r.value) : null);
      })));
  }

  function ProfileFields(props) {
    return h('section', { className: 'mt-fields' },
      h('h3', { className: 'mt-section-label' }, props.label),
      h('dl', null, (props.fields || []).map(function (f, i) {
        return h('div', { key: i }, h('dt', null, f.term), h('dd', null, f.value));
      })),
      props.actionLabel ? h('div', { className: 'mt-fields-actions' }, h(Button, { onClick: props.onAction }, props.actionLabel)) : null);
  }

  /* ---------- TaskWidget: the whole card, every screen ---------- */
  var DEFAULT_STRINGS = {
    profileAction: 'Profile & progress', back: 'Back to task', appName: 'MathTrail', time: 'just now', more: 'More options',
    pick: 'Pick one answer', answers: 'Answers', idk: "I don't know", hint: 'Hint', hideHint: 'Hide hint', another: 'Another task',
    ask: 'Ask a question about the task…', askLabel: 'Ask a question about the task', send: 'Send',
    checking: 'Checking…', yourAnswer: 'Your answer', correctAnswer: 'Correct answer',
    trap: 'The trap', solution: 'Solution', wrongVerdict: "Not quite — it's {correct}, not {picked}.", idkLead: "Here's how to solve it.",
    you: 'You', generatingTitle: 'Preparing the next task…', done: 'Done:', active: 'In progress:', waiting: 'Waiting:',
    rank: 'Rank {rank} of {total}', overall: 'overall rating', nextUp: 'Next up', topics: 'Topics', recent: 'Recent answers',
    mistakes: 'Mistakes that repeat', profile: 'Profile · for the parent', editProfile: 'Edit profile', mastered: 'Mastered', right: 'Right', wrong: 'Wrong',
    taskLabel: 'Task', progressLabel: 'Profile and progress', replies: 'Replies', demoReply: 'In your own chat, the model answers here — about this very task.'
  };

  var DEFAULT_TASK = {
    text: 'A fence is 12 meters long. Posts stand every 3 meters, including both ends. How many posts are there?',
    diagram: '|--3--|--3--|--3--|--3--|',
    options: [{ letter: 'A', value: '3' }, { letter: 'B', value: '4' }, { letter: 'C', value: '5' }, { letter: 'D', value: '6' }, { letter: 'E', value: '12' }],
    correct: 'C',
    hint: 'Try a smaller fence first: 6 meters long, with a post every 3 meters. Draw it and count the posts.',
    traps: { B: 'Counted the gaps instead of the posts.' },
    right: 'Correct! 5 posts.', rightDetail: 'You remembered the posts at both ends.',
    solution: ['12 ÷ 3 = 4 gaps.', 'A straight fence with posts at both ends has one more post than gaps.', '4 + 1 = 5 posts.'],
    question: "why isn't it 6?",
    answer: ["Good question! Let's check 6.", '6 posts, 3 meters apart, have 5 gaps between them: 5 × 3 = 15 meters. But our fence is only 12 meters long.', 'Try drawing the fence with a post at 0, 3, 6… meters, and count them.']
  };

  var DEFAULT_GEN = ['Picked topic and difficulty', 'Writing the task', 'Checking every answer', 'Checking readability for grade 3', 'Ready'];

  function scenePreset(scene, task) {
    var s = { view: 'task', picked: null, checking: false, revealed: null, hintOpen: false, thread: [], gen: 2 };
    switch (scene) {
      case 'generating': s.view = 'generating'; break;
      case 'selected': s.picked = 'B'; s.checking = true; break;
      case 'hint': s.hintOpen = true; break;
      case 'wrong': s.picked = 'B'; s.revealed = 'answer'; break;
      case 'right': s.picked = task.correct; s.revealed = 'answer'; break;
      case 'idk': s.revealed = 'idk'; break;
      case 'question': s.thread = [{ author: 'person', text: task.question }, { author: 'app', paragraphs: task.answer }]; break;
      case 'progress': s.view = 'progress'; break;
      default: break;
    }
    return s;
  }

  function TaskWidget(props) {
    var React = R();
    var t = Object.assign({}, DEFAULT_STRINGS, props.strings || {});
    var task = Object.assign({}, DEFAULT_TASK, props.task || {});
    var child = props.child || { name: 'Comet', badge: 'Olympiad coach · Grade 3', grade: 'Grade 3' };
    var scene = props.scene || 'task';
    var st = React.useState(function () { return scenePreset(scene, task); });
    var s = st[0], set = st[1];
    var timers = React.useRef([]);
    function clearTimers() { timers.current.forEach(clearTimeout); timers.current = []; }
    function later(fn, ms) { timers.current.push(setTimeout(fn, ms)); }
    function patch(p) { set(function (prev) { return Object.assign({}, prev, p); }); }

    React.useEffect(function () { clearTimers(); set(scenePreset(scene, task)); return clearTimers; }, [scene, props.resetKey]);

    // The generating checklist walks forward on its own; in the 'generating' scene it loops.
    React.useEffect(function () {
      if (s.view !== 'generating') return undefined;
      var id = setTimeout(function () {
        if (s.gen < DEFAULT_GEN.length - 1) patch({ gen: s.gen + 1 });
        else if (scene === 'generating' && !s.fromAction) patch({ gen: 0 });
        else set(Object.assign(scenePreset('task', task), {}));
      }, s.gen === DEFAULT_GEN.length - 1 ? 1200 : 900);
      return function () { clearTimeout(id); };
    }, [s.view, s.gen, scene]);

    var answered = !!s.revealed;
    var generating = s.view === 'generating';

    function pick(letter) {
      if (answered || s.checking) return;
      if (props.onPick) props.onPick(letter);
      patch({ picked: letter, checking: true, hintOpen: s.hintOpen });
      later(function () { patch({ checking: false, revealed: 'answer' }); }, 1100);
    }
    function toggleHint() { if (props.onHint) props.onHint(); patch({ hintOpen: !s.hintOpen }); }
    function idk() { if (props.onIdk) props.onIdk(); patch({ revealed: 'idk', checking: false, picked: null }); }
    function another() { clearTimers(); if (props.onAnother) props.onAnother(); set(Object.assign(scenePreset('generating', task), { gen: 0, fromAction: true })); }
    function ask(text) {
      if (props.onAsk) props.onAsk(text);
      var next = s.thread.concat([{ author: 'person', text: text }]);
      patch({ thread: next });
      later(function () {
        var isSix = /\b6\b|шест|six/i.test(text);
        set(function (prev) { return Object.assign({}, prev, { thread: prev.thread.concat([{ author: 'app', paragraphs: isSix ? task.answer : [t.demoReply] }]) }); });
      }, 800);
    }

    function optionState(letter) {
      if (s.revealed === 'answer') {
        if (letter === task.correct) return 'correct';
        if (letter === s.picked) return 'wrong';
        return 'muted';
      }
      if (s.revealed === 'idk') return letter === task.correct ? 'correct' : 'muted';
      if (s.picked === letter && s.checking) return 'selected';
      return 'default';
    }
    function optionStatus(state) {
      return state === 'selected' ? t.checking : state === 'correct' ? t.correctAnswer : state === 'wrong' ? t.yourAnswer : null;
    }

    var wide = !!props.wide;
    var rootClass = cx('mt', 'mt-widget', wide && 'mt-wide', props.dir === 'rtl' && 'mt-rtl', props.className);

    if (s.view === 'progress') {
      var pr = props.progress || {};
      return h('div', { className: rootClass, dir: props.dir || 'ltr' },
        h(ThreadBar, { variant: 'back', label: t.back, onClick: function () { patch({ view: 'task' }); } }),
        h('article', { 'aria-label': t.progressLabel },
          h(MessageHeader, { author: 'person', name: child.name, badge: child.grade, wide: wide, moreLabel: t.more }),
          h('div', { className: 'mt-progress' },
            h(RatingSummary, { rank: pr.rank || 3, total: pr.total || 5, rating: pr.rating || '1573', rankLabel: fmt(t.rank, { rank: pr.rank || 3, total: pr.total || 5 }), ratingLabel: t.overall }),
            h(Note, { tone: 'plain', label: t.nextUp }, pr.next || 'Enumeration again, at the same level'),
            h(StatList, { label: t.topics, rows: pr.topics || [] }),
            h(StatList, { label: t.recent, rows: pr.recent || [] }),
            h(StatList, { label: t.mistakes, rows: pr.mistakes || [] }),
            h(ProfileFields, { label: t.profile, fields: pr.fields || [], actionLabel: t.editProfile }))));
    }

    var body;
    if (generating) {
      body = h('div', { className: 'mt-body', style: { gap: '12px' } },
        h(GeneratingSteps, {
          title: t.generatingTitle,
          statusLabels: { done: t.done, active: t.active, waiting: t.waiting },
          steps: (props.generatingSteps || DEFAULT_GEN).map(function (label, i) {
            var detail = i === 2 ? (props.generatingDetail || '(120 cases, exactly 1 correct)') : null;
            return { label: label, detail: detail, status: i < s.gen ? 'done' : i === s.gen ? (i === DEFAULT_GEN.length - 1 ? 'done' : 'active') : 'waiting' };
          })
        }));
    } else {
      body = h('div', { className: 'mt-body' },
        h('p', { className: 'mt-task-text' }, task.text),
        task.diagram ? h(Diagram, null, task.diagram) : null,
        s.hintOpen && !answered ? h(Note, { tone: 'hint', label: t.hint }, task.hint) : null,
        h(OptionList, {
          legend: answered ? t.answers : t.pick,
          name: 'mt-answer',
          disabled: answered,
          onSelect: pick,
          options: task.options.map(function (o) {
            var state = optionState(o.letter);
            return { letter: o.letter, value: o.value, state: state, status: optionStatus(state) };
          })
        }));
    }

    var replies = [];
    if (s.revealed === 'answer') {
      var right = s.picked === task.correct;
      var pickedOpt = task.options.filter(function (o) { return o.letter === s.picked; })[0] || {};
      var correctOpt = task.options.filter(function (o) { return o.letter === task.correct; })[0] || {};
      replies.push(h(ReplyCard, { key: 'verdict', author: 'app', name: t.appName, time: t.time, moreLabel: t.more },
        right ? h(Verdict, { tone: 'correct', detail: task.rightDetail }, task.right)
          : h(Verdict, { tone: 'wrong' }, fmt(t.wrongVerdict, { correct: correctOpt.value, picked: pickedOpt.value })),
        !right && task.traps[s.picked] ? h(Note, { tone: 'trap', label: t.trap }, task.traps[s.picked]) : null,
        h(SolutionSteps, { label: t.solution, steps: task.solution })));
    } else if (s.revealed === 'idk') {
      replies.push(h(ReplyCard, { key: 'idk', author: 'app', name: t.appName, time: t.time, moreLabel: t.more },
        h(Verdict, { tone: 'none' }, t.idkLead),
        h(SolutionSteps, { label: t.solution, steps: task.solution })));
    }
    s.thread.forEach(function (m, i) {
      if (m.author === 'person') {
        replies.push(h(ReplyCard, { key: 'q' + i, author: 'person', name: t.you, time: t.time, moreLabel: t.more },
          h('p', { className: 'mt-reply-lead' }, m.text)));
      } else {
        replies.push(h(ReplyCard, { key: 'a' + i, author: 'app', name: t.appName, time: t.time, tight: true, moreLabel: t.more },
          (m.paragraphs || []).map(function (p, j) { return h('p', { key: j }, p); })));
      }
    });

    var buttons = answered
      ? [h(Button, { key: 'another', variant: 'primary', onClick: another }, t.another)]
      : [
        h(Button, { key: 'idk', disabled: generating, onClick: idk }, t.idk),
        h(Button, { key: 'hint', disabled: generating, pressed: s.hintOpen, onClick: toggleHint }, s.hintOpen ? t.hideHint : t.hint),
        h(Button, { key: 'another', disabled: generating, onClick: another }, t.another)
      ];

    return h('div', { className: rootClass, dir: props.dir || 'ltr' },
      h(ThreadBar, { name: child.name, action: t.profileAction, onClick: function () { if (props.onProfile) props.onProfile(); patch({ view: 'progress' }); } }),
      h('article', { 'aria-label': t.taskLabel },
        h(MessageHeader, { author: 'app', name: t.appName, time: t.time, badge: child.badge, wide: wide, moreLabel: t.more }),
        body),
      replies.length ? h('section', { className: 'mt-replies', 'aria-label': t.replies, 'aria-live': 'polite' }, replies) : null,
      h('div', { className: 'mt-foot' },
        h(ReplyField, { placeholder: t.ask, label: t.askLabel, sendLabel: t.send, disabled: generating, onSend: ask }),
        h('div', { className: 'mt-btns' }, buttons)));
  }

  var api = {
    TaskWidget: TaskWidget, OptionRow: OptionRow, OptionList: OptionList, Button: Button, ReplyField: ReplyField,
    ThreadBar: ThreadBar, MessageHeader: MessageHeader, Badge: Badge, Diagram: Diagram, Note: Note, Verdict: Verdict,
    SolutionSteps: SolutionSteps, ReplyCard: ReplyCard, GeneratingSteps: GeneratingSteps, RatingSummary: RatingSummary,
    StatList: StatList, ProfileFields: ProfileFields, Icon: Icon, Mark: Mark, Avatar: Avatar
  };
  window.MathTrail = Object.assign(window.MathTrail || {}, api);
})();
