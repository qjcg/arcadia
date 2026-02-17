#import "diploma.typ": diploma

#let student = sys.inputs.at("student", default: "Student Name")
#let course = sys.inputs.at("course", default: "Course Name")
#let instructor = sys.inputs.at("instructor", default: "Instructor Name")
#let dates = sys.inputs.at("dates", default: "Dates")

#diploma(
  student: student,
  course: course,
  instructor: instructor,
  dates: dates,
)
