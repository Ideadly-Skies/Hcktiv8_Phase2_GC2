-- fetches all students, their contact info, and the courses they are currently enrolled in
SELECT 
    s.id AS student_id,
    s.first_name,
    s.last_name,
    s.email,
    s.address,
    c.id AS course_id,
    c.name AS course_name
FROM 
    students s
LEFT JOIN 
    enrollments e ON s.id = e.student_id
LEFT JOIN 
    courses c ON e.course_id = c.id;

-- retrieves all courses, their respectives departments, professors teaching them, and students enrolled
SELECT 
    c.id AS course_id,
    c.name AS course_name,
    d.name AS department_name,
    p.first_name AS professor_first_name,
    p.last_name AS professor_last_name,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name
FROM 
    courses c
LEFT JOIN 
    departments d ON c.department_id = d.id
LEFT JOIN 
    teachings t ON c.id = t.course_id
LEFT JOIN 
    professors p ON t.professor_id = p.id
LEFT JOIN 
    enrollments e ON c.id = e.course_id
LEFT JOIN 
    students s ON e.student_id = s.id;

-- lists all professors with their contact information and the courses they teach
SELECT 
    p.id AS professor_id,
    p.first_name,
    p.last_name,
    p.email,
    p.address,
    c.id AS course_id,
    c.name AS course_name
FROM 
    professors p
LEFT JOIN 
    teachings t ON p.id = t.professor_id
LEFT JOIN 
    courses c ON t.course_id = c.id;

-- provides enrollments dates and credits for each student's enrollment in courses
SELECT 
    e.student_id,
    e.course_id,
    e.enrollment_date,
    c.credits
FROM 
    enrollments e
LEFT JOIN 
    courses c ON e.course_id = c.id;

-- lists all departments and the courses that belong to them
SELECT 
    d.id AS department_id,
    d.name AS department_name,
    c.id AS course_id,
    c.name AS course_name
FROM 
    departments d
LEFT JOIN 
    courses c ON d.id = c.department_id;

-- counts the total number of students enrolled in each course
SELECT 
    c.id AS course_id,
    c.name AS course_name,
    COUNT(e.student_id) AS total_students
FROM 
    courses c
LEFT JOIN 
    enrollments e ON c.id = e.course_id
GROUP BY 
    c.id, c.name;

-- calculates the average number of students enrolled in courses for each department
SELECT 
    d.id AS department_id,
    d.name AS department_name,
    AVG(student_count) AS average_students
FROM 
    (
        SELECT 
            c.department_id,
            COUNT(e.student_id) AS student_count
        FROM 
            courses c
        LEFT JOIN 
            enrollments e ON c.id = e.course_id
        GROUP BY 
            c.department_id, c.id
    ) AS department_student_counts
LEFT J
