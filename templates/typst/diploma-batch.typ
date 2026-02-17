#import "diploma.typ": diploma

#let data = yaml("diploma-students.yaml")

#for student in data.students {
  diploma(
    student: student,
    course: data.course,
    instructor: data.instructor,
    dates: data.dates,
  )
}
