// quiz.js — reusable quiz component for Pavona lessons.
//
// Usage (see lessons for worked examples):
//   <div class="quiz-q">
//     <p class="prompt">Question text?</p>
//     <div class="quiz-options">
//       <button class="quiz-option">Answer A</button>
//       <button class="quiz-option" data-correct>Answer B</button>
//       ...
//     </div>
//     <div class="quiz-feedback"><span class="verdict"></span> Explanation.</div>
//   </div>
//
// Clicking an option gives immediate feedback: the correct option is highlighted,
// the clicked wrong option is marked, and the explanation is revealed. Questions
// are single-attempt — retrieval practice, not guessing until right.

document.addEventListener("click", function (event) {
  const option = event.target.closest(".quiz-option");
  if (!option || option.disabled) return;

  const question = option.closest(".quiz-q");
  if (!question) return;

  const options = question.querySelectorAll(".quiz-option");
  const feedback = question.querySelector(".quiz-feedback");
  const verdict = feedback ? feedback.querySelector(".verdict") : null;
  const isCorrect = option.hasAttribute("data-correct");

  options.forEach(function (btn) {
    btn.disabled = true;
    if (btn.hasAttribute("data-correct")) btn.classList.add("correct");
  });
  if (!isCorrect) option.classList.add("wrong");

  if (feedback) {
    if (verdict) verdict.textContent = isCorrect ? "✓ Correct. " : "✗ Not quite. ";
    feedback.classList.add("show");
  }
});
