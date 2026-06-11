#import "letter.typ": letter

#show: letter.with(
  from-name: "Arcadia Foundation",
  from-address: (
    "742 Evergreen Terrace",
    "Springfield, IL 62701",
    "United States",
  ),
  from-email: "grants@arcadia.foundation",
  from-phone: "+1 (555) 849-2731",
  to-name: "Dr. Helena Whitcroft",
  to-title: "Director of Research",
  to-company: "Northern Institute of Technology",
  to-address: (
    "100 University Avenue",
    "Cambridge, MA 02139",
    "United States",
  ),
  date: datetime(year: 2026, month: 3, day: 14),
  subject: "Notice of Grant Award — Project Hermes",
  salutation: "Dear Dr. Whitcroft,",
  closing: "With warm regards,",
  signatory: "Marcus Aurelius Vega",
  signatory-title: "Chief Grants Officer, Arcadia Foundation",
  enclosures: ("Grant Award Terms & Conditions (signed)", "Payment Schedule Overview", "Reporting Guidelines"),
)

We are delighted to inform you that your proposal titled *"Project Hermes: Next-Generation Optical Communication Networks"* has been selected for funding by the Arcadia Foundation Review Board.

#let award-color = rgb("#2c3e50")
#align(center, text(size: 24pt, weight: "bold", fill: award-color, [\$400,000]))
#align(center, text(size: 10.5pt, fill: gray, [Awarded over 24 months]))

After a rigorous evaluation by an independent panel of experts, your application was unanimously recommended for its scientific merit, feasibility, and potential societal impact. The selection committee was particularly impressed with the interdisciplinary scope of your methodology.

The grant will be disbursed in quarterly installments beginning *April 1, 2026*, subject to the terms outlined in the enclosed agreement. A signed copy of the Grant Award Terms & Conditions must be returned to our office by *April 7, 2026* to activate the funding.

We look forward to following the progress of Project Hermes and welcome the opportunity to support your outstanding research. Should you have any questions, please do not hesitate to contact your program officer, Dr. Yuki Tanaka, at y.tanaka\@arcadia.foundation.