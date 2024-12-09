DROP TABLE IF EXISTS teachings;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS professors;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS students;

-- Create the students table
CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    address TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    jwt_token TEXT
);

-- Create the departments table
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL
);

-- Create the courses table
CREATE TABLE courses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    department_id INT NOT NULL,
    credits INT NOT NULL,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE
);

-- Create the professors table
CREATE TABLE professors (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    address TEXT NOT NULL
);

-- Create the enrollments table
CREATE TABLE enrollments (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    course_id INT NOT NULL,
    enrollment_date DATE NOT NULL,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    UNIQUE(student_id, course_id)
);

-- Create the teachings table
CREATE TABLE teachings (
    id SERIAL PRIMARY KEY,
    professor_id INT NOT NULL,
    course_id INT NOT NULL,
    FOREIGN KEY (professor_id) REFERENCES professors(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    UNIQUE(professor_id, course_id)
);

-- Seed data for the students table
INSERT INTO students (first_name, last_name, email, password_hash, address, date_of_birth, jwt_token) VALUES
('Obie', 'Ananda', 'obie.ananda@example.com', 'hashedpassword1', 'Jl. Pare 1 Blok D6/No.18, BSD City, Tangerang Selatan', '2001-10-23', 'jwt_token_1'),
('Rachel', 'Kwok', 'rachel.kwok@maths.usyd.edu', 'hashedpassword2', 'Surry Hills, New South Wales, Australia', '2002-09-22', 'jwt_token_2'),
('Devan', 'Ananda', 'devan.ananda@example.com', 'hashedpassword3', '8 Somapah Rd, #4th Floor Building 1, Singapore 487372', '2004-06-15', 'jwt_token_3');

-- Seed data for the departments table
INSERT INTO departments (name, description) VALUES
('Computer Science', 'The Computer Science department focuses on software development, data analysis, and computing systems.'),
('Mathematics', 'The Mathematics department emphasizes theoretical and applied mathematics.'),
('Physics', 'The Physics department focuses on the study of matter, energy, and the universe.');

-- Seed data for the courses table
INSERT INTO courses (name, description, department_id, credits) VALUES
('Object Oriented Programming', 'An introduction into the world of object-oriented programming using java', 1, 4),
('Calculus', 'An in-depth course on differential and integral calculus.', 2, 3),
('Quantum Mechanics', 'An advanced course on quantum physics principles.', 3, 4);

-- Seed data for the professors table
INSERT INTO professors (first_name, last_name, email, address) VALUES
('John', 'Stravrakakis', 'john.stavrakakis@sydney.edu.au', 'Sydney, NSW, Australia'),
('Sarah', 'Taylor', 'sarah.taylor@america.com', '34 College Ln, Townsville'),
('David', 'Lee', 'david.lee@america.com', '56 Campus Way, Villageton');

-- Seed data for the enrollments table
INSERT INTO enrollments (student_id, course_id, enrollment_date) VALUES
(1, 1, '2023-09-01'),
(2, 2, '2023-09-01'),
(3, 3, '2023-09-01'),
(1, 2, '2023-09-15'),
(2, 3, '2023-09-15');

-- Seed data for the teachings table
INSERT INTO teachings (professor_id, course_id) VALUES
(1, 1),
(2, 2),
(3, 3),
(1, 2),
(3, 1);