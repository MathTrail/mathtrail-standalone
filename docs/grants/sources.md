# The works the logic model rests on

Each work below was read in full on 5 October 2026, from the open copy it names, and its metadata were checked against Crossref by its DOI, corrections and retractions included. A passage is copied word for word from that copy, and a page is the copy's own, or a section where the copy has no pages. Every entry says which key component of the [logic model](logic-model.md) the work supports, and how far its learners are from ours: children in grades 1–6, solving olympiad-style problems in a chat with an adult beside them.

Five distinct empirical works are counted. A sixth stands beside the first, because it reports one of its trials, and a seventh waits for a copy that can be read.

## S1. Adaptive practice held in a band of success

Pelánek, R., Papoušek, J., Řihák, J., Stanislav, V., & Nižnan, J. (2017). Elo-based learner modeling for the adaptive practice of facts. *User Modeling and User-Adapted Interaction*, 27, 89–118. https://doi.org/10.1007/s11257-016-9185-7

- **Copy read.** The authors' manuscript, 33 pages: https://www.fi.muni.cz/~xpelanek/publications/umuai-adaptive-practice.pdf
- **Design.** Two randomized trials in an open adaptive system for practising geography facts, used mostly by Czech students. The first compared adaptive with random construction of questions (August to October 2015, more than 1,300,000 answers from about 20,000 learners). The second compared target success rates of 50, 65, 80 and 95 per cent (November 2015 to January 2016, almost 3,300,000 answers from about 37,000 learners). On about 3,900,000 answers of 91,000 learners, an Elo learner model estimated item difficulty nearly as joint maximum likelihood does (correlation 0.97), in one pass and online.
- **Finding.** Adaptive construction helped both engagement and learning. More difficult practice led to better learning, the difference lying mainly between the easiest condition and the others. Learners rated practice at 60–70 per cent success the most appropriate, and those using the system at school preferred about 75 per cent.
- **Passages.**
  - "The results show that the adaptive behavior is beneficial (both for engagement and learning) and indicate which aspects of adaptivity are important – adaptive choice of the number of distractors rather than the choice of the target item." (p. 28)
  - "The results of our experiments suggest that a suitable success rate is around 65 %." (p. 28)
  - "We have also detected differences between in-school and out-of-school usage: students using the system in schools prefer easier questions, which accords with previous literature (Abuhamdeh and Csikszentmihalyi, 2012)." (p. 28)
- **Supports.** C1: a learner model of the Elo family holds practice in a band of likely success.
- **How far from ours.** Facts to remember, not problems to solve; learners of unspecified age practising by choice; and the band school learners preferred lies higher than the one the trial favoured.

Beside it: Papoušek, J., Stanislav, V., & Pelánek, R. (2016). Impact of question difficulty on engagement and learning. In *Intelligent Tutoring Systems*, LNCS 9684, 267–272. https://doi.org/10.1007/978-3-319-39583-8_28 — the short paper of the second trial, read in the authors' copy (https://www.fi.muni.cz/~xpelanek/publications/its-target-difficulty.pdf): "The experiment shows that easy questions are better for short term engagement, whereas difficult questions are better for long term engagement and learning." (p. 1). It is one trial with S1, so it is not counted apart.

## S2. A model's mathematics has to be checked

Pardos, Z. A., & Bhandari, S. (2024). ChatGPT-generated help produces learning gains equivalent to human tutor-authored help on mathematics skills. *PLOS ONE*, 19(5), e0304013. https://doi.org/10.1371/journal.pone.0304013

- **Copy read.** The published article, open under CC BY, 18 pages: https://journals.plos.org/plosone/article?id=10.1371/journal.pone.0304013
- **Design.** A randomized 3 × 4 online experiment with 274 adults (Mechanical Turk, median age 36–40) in four areas of algebra and statistics: worked-solution hints written by ChatGPT 3.5, hints written by human tutors, or no hints, with a pre-test and a repeated post-test.
- **Finding.** ChatGPT's hints produced a significant gain (17 per cent) over the control (1.85 per cent), not separable from the human tutors' (11.62 per cent). But 32 per cent of the hints ChatGPT wrote failed the quality checks with a wrong answer and wrong work; asking ten times and keeping the most common answer cut that to nearly 0 per cent in algebra and 13 per cent in statistics.
- **Passages.**
  - "Notably, ChatGPT-generated help failed quality checks on 32% of problems." (p. 1)
  - "This all suggests that ChatGPT and other LLMs should not be used to give feedback to students in the same way a teacher or teaching assistant would, unless it was in a domain verified to have near zero error." (p. 13)
- **Supports.** C3: a program checks every task the model writes before the child sees it.
- **How far from ours.** Adults, not children; hints to textbook problems, not whole problems written afresh; and the model of December 2022.

## S3. Without guardrails a chatbot harms learning

Bastani, H., Bastani, O., Sungu, A., Ge, H., Kabakcı, Ö., & Mariman, R. (2025). Generative AI without guardrails can harm learning: Evidence from high school mathematics. *Proceedings of the National Academy of Sciences*, 122(26), e2422633122. https://doi.org/10.1073/pnas.2422633122 (corrected for one author's affiliation alone: https://doi.org/10.1073/pnas.2518204122)

- **Copy read.** The published article's full text in Europe PMC (PMC12232635), open under CC BY-NC-ND; its supplement was not read.
- **Design.** A preregistered randomized controlled trial with nearly 1,000 students in grades 9–11 at a high school in Turkey, over four 90-minute sessions. Classes practised with a ChatGPT-like tutor ("GPT Base"), with a tutor holding the teachers' solutions and common mistakes and told to give hints rather than answers ("GPT Tutor"), or with books and notes alone, and then sat an exam on their own.
- **Finding.** Practice scores rose by 48 per cent with GPT Base and 127 per cent with GPT Tutor. On the exam that followed, the GPT Base group did 17 per cent worse than the control, while the GPT Tutor group did not differ from it. Asked for answers, GPT Base was right only 51 per cent of the time.
- **Passages.**
  - "Second, on the exam, students in the GPT Base arm perform statistically significantly worse than students in the control arm by 17%; this negative effect is essentially eradicated in the GPT Tutor arm, though we still do not observe a positive effect." (Introduction)
  - "First, the prompt instructs GPT-4 to provide hints to the student without directly giving them the answer, to encourage learning ( 20 ). Second, the prompt provides a significant amount of problem-specific information provided by teachers, including one or more (correct) solutions to the practice problem, as well as common student mistakes and how to provide feedback." (Experimental Design)
  - "We find that GPT Base gives a correct answer only 51% of the time on average; it makes logical errors 42% of the time and arithmetic errors 8% of the time." (Potential Mechanism: Asking for Solutions)
- **Supports.** C4: the answer stays sealed until the child answers, and a hint comes only when asked for. Also C3, the checked solution, and C5, feedback built on known mistakes.
- **How far from ours.** Older students; solutions written by teachers for every problem; one school, and the models of autumn 2023.

## S4. Feedback right after the answer helps young children

Fyfe, E. R., & Rittle-Johnson, B. (2016). The benefits of computer-generated feedback for mathematics problem solving. *Journal of Experimental Child Psychology*, 147, 140–151. https://doi.org/10.1016/j.jecp.2016.03.009

- **Copy read.** The accepted manuscript in ERIC (ED566264), 28 pages: https://files.eric.ed.gov/fulltext/ED566264.pdf
- **Design.** A randomized experiment with 75 second-grade children (mean age 8.2), who were taught a strategy for mathematical equivalence problems and then solved twelve on a computer, with no feedback, with the correct answer after each problem, or with all the correct answers at the end; a posttest followed the next day.
- **Finding.** On the next-day posttest, children scored 86 per cent with feedback after each problem, 78 per cent with feedback at the end and 65 per cent with none. Feedback helped the children with low prior knowledge; only feedback after each problem raised mastery (63 per cent solved every item, against 40 and 38 per cent) and transfer to new problems, and it cut the use of the common wrong strategies. Feedback at the end lowered the children's assessment of themselves.
- **Passages.**
  - "Immediate feedback was particularly effective, facilitating mastery of the material for children with both low and high prior knowledge." (p. 2)
  - "That is, when children in the immediate feedback condition made errors, they were less likely to use common incorrect strategies that stem from entrenched misconceptions (McNeil & Alibali, 2005) and more likely to try something different." (p. 20)
- **Supports.** C5: the mistake behind a wrong answer is explained right after the answer, about the task and not the child.
- **How far from ours.** The feedback was the correct answer alone, after a lesson on the strategy; MathTrail explains the trap and the solution, on olympiad-style problems.

## S5. Mastery is estimated, and the estimate runs ahead

Corbett, A. T., & Anderson, J. R. (1995). Knowledge tracing: Modeling the acquisition of procedural knowledge. *User Modeling and User-Adapted Interaction*, 4, 253–278. https://doi.org/10.1007/BF01099821

- **Copy read.** The authors' scan hosted by the ACT-R group at Carnegie Mellon, 26 pages, without a text layer and read through optical character recognition: http://act-r.psy.cmu.edu/wordpress/wp-content/uploads/2012/12/893CorbettAnderson1995.pdf
- **Design.** Four studies of knowledge tracing in the ACT Programming Tutor, with college students learning Lisp. In the fourth, 25 students practised until the tutor estimated they had mastered every rule, and 21 completed only the 38 required exercises.
- **Finding.** 56 per cent of the students who practised to mastery reached 90 per cent on the test, against 24 per cent of the others. The estimate ran ahead of the students: their actual test accuracy was 5–10 per cent lower than predicted, and a rule mastered in simple settings could fail in harder ones.
- **Passages.**
  - "If we adopt a strict operational definition of mastery as 90% correct on the test, Figure 7 indicates that 56% of the students in this knowledge tracing condition reached the mastery performance level. A second group of twenty-one students in this study who completed just the 38 required exercises provides a contrast. Only 24% percent of these students reached this high level of test performance." (p. 274)
  - "Students may get credit for having mastered an “ideal” rule in simple contexts that, in fact, does not transfer to more complex contexts." (p. 266)
  - "Average actual and predicted accuracy are again displayed in the first two columns of the table and actual accuracy is 5%—10% lower than predicted." (p. 273)
- **Supports.** C6: a topic counts as mastered cautiously, by an estimate, and the mastery can be taken back.
- **How far from ours.** College students learning to program in a tutor; a two-state model of each rule, with mastery at 0.95.

## Waiting for a copy

Rebholz, F., Golle, J., Tibus, M., Ruth-Herbein, E., Moeller, K., & Trautwein, U. (2022). Getting fit for the Mathematical Olympiad: positive effects on achievement and motivation? *Zeitschrift für Erziehungswissenschaft*, 25, 1175–1198. https://doi.org/10.1007/s11618-022-01106-y

An open-access quasi-experiment with children in grades 3–4 taking a course for the Mathematical Olympiad, and the work closest to C2. The publisher and both repositories that hold it refuse automated readers, so it is read once a copy is downloaded by hand; until then C2 rests on no work read.

## Not used, and why

- **The memo's candidates.** Dunlosky and colleagues (2013), Hattie and Timperley (2007), Bjork and Bjork (2011) and Pelánek (2016) are reviews, which inform the model but are not counted as empirical works. Sweller and Cooper (1985) is empirical, but no component of MathTrail as it stands rests on worked examples studied first.
- **Math Garden.** Klinkenberg, Straatemeier and van der Maas (2011), adaptive arithmetic practice for children with an Elo model, is the closest system to ours, but no open copy of it could be found.
