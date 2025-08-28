-- ===========================
-- DROP TABLES IN DEPENDENCY ORDER
-- ===========================

DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS sprints CASCADE;
DROP TABLE IF EXISTS project_has_users CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ===========================
-- RECREATE SCHEMA + DATA
-- ===========================

CREATE TABLE users (
  id UUID NOT NULL,
  username VARCHAR(30) NOT NULL,
  password VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL,
  first_name VARCHAR(50) NOT NULL,
  last_name VARCHAR(50) NOT NULL,
  active BOOLEAN NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE password_resets (
  id SERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  reset_token VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_password_resets_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_password_resets_reset_token ON password_resets (reset_token);
CREATE INDEX idx_password_resets_user_id ON password_resets (user_id);

CREATE TABLE projects (
  id UUID NOT NULL,
  name VARCHAR(50) NOT NULL,
  description VARCHAR(255) NOT NULL,
  project_owner_id UUID NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_projects_owner_id FOREIGN KEY (project_owner_id) REFERENCES users (id)
);

CREATE INDEX idx_projects_owner_id ON projects (project_owner_id);

CREATE TABLE project_has_users (
  project_id UUID NOT NULL,
  user_id UUID NOT NULL,
  CONSTRAINT fk_project_has_users_project_id FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
  CONSTRAINT fk_project_has_users_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_has_users_project_id ON project_has_users (project_id);
CREATE INDEX idx_project_has_users_user_id ON project_has_users (user_id);

CREATE TABLE sprints (
  id UUID NOT NULL,
  name VARCHAR(50) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  end_date TIMESTAMP NOT NULL,
  is_completed BOOLEAN NOT NULL,
  is_started BOOLEAN NOT NULL,
  project_id UUID NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_sprints_project_id FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);

CREATE INDEX idx_sprints_project_id ON sprints (project_id);

CREATE TABLE tasks (
  id UUID NOT NULL,
  description VARCHAR(255) NOT NULL,
  assignee_id UUID DEFAULT NULL,
  date_created TIMESTAMP NOT NULL,
  status INTEGER NOT NULL,
  sprint_id UUID NOT NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_tasks_assignee_id FOREIGN KEY (assignee_id) REFERENCES users (id) ON DELETE SET NULL,
  CONSTRAINT fk_tasks_sprint_id FOREIGN KEY (sprint_id) REFERENCES sprints (id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX idx_tasks_sprint_id ON tasks (sprint_id);

-- Insert users data
INSERT INTO users (id, username, password, email, first_name, last_name, active) VALUES
('03abaf8e-9801-43be-bf90-8f610e5a1137', 'Kaelen', '$2a$11$sK7X4s6Id5ciKwbAHPrbpuxstiIcPzrOM4m2duCdu/2zUaVAUg.ku', 'Kaelen@gmail.com', 'Kaelen', 'Virelli', true),
('0e98413c-0078-4bdb-be9c-b31c6c7097b0', 'Karsten', '$2a$11$IlJLgFdcZwsRgOpV429Qhe1C941M/5yDhghmkOJGEb.x2yjB1nxSO', 'karstenhedrick@icloud.com', 'Karsten', 'Hedrick', true),
('1678d1ca-5f4d-47e4-bf58-1471450cb27d', 'testtest', '$2a$11$cIK3UCZBOrwnYF7G3fqfVuMWaLWv4C3Vdxl8T5bxtGhSmCZtI44dq', 'test@test2.com', 'test', 'test', true),
('25a0df35-859c-4f5c-8551-dfa12e1e0409', 'test2', '$2a$11$3KtbVACQfCpdGDddtChlqu5NudkU5tZc4Gwqpj3kFhB0lYHei3Hky', 'test2@test', 'test', 'test', false),
('486b539c-33d5-4518-8295-6019b1dc8816', 'Rhea', '$2a$11$OqGmTmFfxpMI6BsGoN5paevnUqACGH8yT2CUELodWfpC4zUYTwEH6', 'Rhea@gmail.com', 'Rhea', 'Solvane', true),
('6d3c8063-1033-4d49-99c4-a5251327d4c0', 'DevHiveUser', '$2a$11$OMh0/n7GP2pIALewre5a8ejjRnIHgR6FGmlndb304o3GpV4LYggS2', 'DevHiveUser@DevHiveUser.com', 'John', 'Doe', true),
('75880697-d348-4d2c-92c9-af2bc7ca19b9', 'kaihedrick', '$2a$11$1kyUcRrjAKeEkZXUXGNs/eeCY/oBE03sHBQG9P4flBd00uupgTLm2', 'kaihedrick@me.com', 'kai', 'hedrick', false),
('ae20138d-ceaa-445c-865c-17ecd5a1059b', 'ka', '$2a$11$Vi12nFEUUFAzFU85/yOLjuLVsf6B/Cpu6VjY5tPkBFob6Dv0/6iQK', 'ka@ka.com', 'ka', 'kai', false),
('c9c3efc6-7ed8-4e7b-b7e3-b937a6df0f8c', 'kevinklark', '$2a$11$tzyjnF4XgPso.LFYPunTfeiL9kGRbj3WlaTl32Vgm0DsFd1Q1JOYy', 'kevinklark@gmail.com', 'kevin', 'klark', false),
('d7e22cfb-7baa-4bae-84c1-09a8d6937cd5', 'johndoe', '$2a$11$5w8umCSQ8svY9DZfA0kQzeIo70hhz293FVRvc/Jo.kQInnnCoFFde', 'johndoe@gmail.com', 'john', 'doe', false),
('de2720d6-4758-4a4b-904b-a339fc2b2187', 'kai', '$2a$11$paOPGHicu3jDx6Dih/MFVOAGadHwnLW6mb0gc2KNPgyTRLKmfvSB6', 'kaihedrick@icloud.com', 'Kai', 'Hedrick', false);

-- Insert password_resets data
INSERT INTO password_resets (id, user_id, reset_token, expires_at, created_at) VALUES
(1, 'de2720d6-4758-4a4b-904b-a339fc2b2187', 'uhUbPSxe20yl4d5sOY93Pw==', '2025-03-21 14:50:33', '2025-03-21 14:20:32'),
(2, 'de2720d6-4758-4a4b-904b-a339fc2b2187', 'CBSAgRehkU+ReqE7E9bXxg==', '2025-03-21 17:18:22', '2025-03-21 16:48:21'),
(6, '75880697-d348-4d2c-92c9-af2bc7ca19b9', 'ApbVVKOYXEidVNPpI40K3A==', '2025-03-23 02:01:46', '2025-03-23 01:31:45'),
(7, '75880697-d348-4d2c-92c9-af2bc7ca19b9', 'xIB6WRlAiUehJkt4bqJ7Wg==', '2025-03-23 02:04:12', '2025-03-23 01:34:11'),
(8, '75880697-d348-4d2c-92c9-af2bc7ca19b9', 'cxOszjiVjUWRlrWFXGqPew==', '2025-03-23 02:05:43', '2025-03-23 01:35:43'),
(11, '1678d1ca-5f4d-47e4-bf58-1471450cb27d', '5rPMQqQAxkKULR8fv49QNw==', '2025-03-23 20:15:41', '2025-03-23 19:45:41'),
(12, 'de2720d6-4758-4a4b-904b-a339fc2b2187', 'zk2VR8IqNk2IaLvI6wmg8g==', '2025-03-23 20:16:05', '2025-03-23 19:46:05'),
(13, 'de2720d6-4758-4a4b-904b-a339fc2b2187', 'wi2O+qv/hkG7k6lCVH4kMw==', '2025-03-23 20:16:41', '2025-03-23 19:46:40');

-- Insert projects data
INSERT INTO projects (id, name, description, project_owner_id) VALUES
('4158da45-27fd-420b-b190-efc5acd57c90', 'NovaCore', 'NovaCore is a modern, next-generation platform designed to serve as the central hub for digital innovation and enterprise solutions.', 'de2720d6-4758-4a4b-904b-a339fc2b2187'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', 'NeuraSync', 'NeuraSync is an intelligent collaboration platform designed to streamline team workflows using AI-driven task suggestions, adaptive timelines, and real-time communication across web and mobile.', '0e98413c-0078-4bdb-be9c-b31c6c7097b0');

-- Insert project_has_users data
INSERT INTO project_has_users (project_id, user_id) VALUES
('4158da45-27fd-420b-b190-efc5acd57c90', 'de2720d6-4758-4a4b-904b-a339fc2b2187'),
('4158da45-27fd-420b-b190-efc5acd57c90', '03abaf8e-9801-43be-bf90-8f610e5a1137'),
('4158da45-27fd-420b-b190-efc5acd57c90', '486b539c-33d5-4518-8295-6019b1dc8816'),
('4158da45-27fd-420b-b190-efc5acd57c90', '0e98413c-0078-4bdb-be9c-b31c6c7097b0'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', '0e98413c-0078-4bdb-be9c-b31c6c7097b0'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', 'de2720d6-4758-4a4b-904b-a339fc2b2187'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', '486b539c-33d5-4518-8295-6019b1dc8816'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', '03abaf8e-9801-43be-bf90-8f610e5a1137'),
('4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd', 'c9c3efc6-7ed8-4e7b-b7e3-b937a6df0f8c');

-- Insert sprints data
INSERT INTO sprints (id, name, start_date, end_date, is_completed, is_started, project_id) VALUES
('15790a02-cc00-43b7-9ae8-8a728b1d35ca', 'Build', '2025-04-03 00:00:00', '2025-04-06 00:00:00', false, false, '4158da45-27fd-420b-b190-efc5acd57c90'),
('212881d5-e892-4ebd-9193-de7f22c55a2f', 'Focus', '2025-04-08 00:00:00', '2025-04-14 00:00:00', false, false, '4158da45-27fd-420b-b190-efc5acd57c90'),
('5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8', 'UI Overhaul', '2025-04-30 00:00:00', '2025-05-15 00:00:00', false, false, '4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd'),
('65041d64-f69f-48cb-9b23-8b8796debac7', 'MVP Build', '2025-04-09 00:00:00', '2025-04-22 00:00:00', false, true, '4cb4c3e0-b35f-400a-8fde-6a248ba6f0dd'),
('695e8d05-b68a-428d-8a9f-2a4585fc5d93', 'Move Forward', '2025-04-16 00:00:00', '2025-05-20 00:00:00', false, false, '4158da45-27fd-420b-b190-efc5acd57c90'),
('cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5', 'Launch', '2025-03-31 00:00:00', '2025-04-02 00:00:00', false, true, '4158da45-27fd-420b-b190-efc5acd57c90');

-- Insert tasks data
INSERT INTO tasks (id, description, assignee_id, date_created, status, sprint_id) VALUES
('00acba00-11f1-4f6f-9f0f-20fd949040c0', 'Set up deployment target (Render or Heroku)', NULL, '2025-03-31 01:07:20', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('08af5126-01a2-4d15-bed8-0fe3399f8016', 'Polish modal spacing and animation transitions', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-04-09 08:15:26', 2, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('18fb98f2-943e-49d6-896b-76445f0902bd', 'Implement responsive grid layout for Kanban board', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-04-09 08:14:27', 0, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('1a924210-c100-41e3-b7b7-723ac396b713', 'Create task model and controller', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-04-09 08:10:01', 0, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('218ed5cb-d811-40ad-86f5-c2a91a673539', 'Identify accessibility improvements', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-04-09 08:14:09', 0, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('33341a28-e3d8-4f39-9ff5-7d6cc519f824', 'Connect Spring Boot app to MySQL/PostgreSQL', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-03-31 01:08:49', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('3bf4afc2-b0a6-4d2a-8986-0718b0024e69', 'Create user settings panel', '0e98413c-0078-4bdb-be9c-b31c6c7097b0', '2025-04-09 08:09:50', 0, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('44dd64a8-1129-477a-a5a8-c55ab51bd8ca', 'Update button styles and hover interactions', '0e98413c-0078-4bdb-be9c-b31c6c7097b0', '2025-04-09 08:14:57', 0, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('59d23f2f-7ba2-4c0d-84e7-2b374b46ca2f', 'Implement task filters', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-04-09 08:10:23', 1, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('773106c7-df57-4049-8c23-7861c1f4b2f3', 'Redesign login/register forms with icon alignment', 'c9c3efc6-7ed8-4e7b-b7e3-b937a6df0f8c', '2025-04-09 08:14:39', 1, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('7bdef83e-2042-4ee3-ab66-53fa579cac4e', 'Design landing page', 'c9c3efc6-7ed8-4e7b-b7e3-b937a6df0f8c', '2025-04-09 08:10:31', 0, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('7e711136-025c-45b2-9432-7b72d1eb5530', 'Created User entity/model and repository', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-03-31 01:09:24', 2, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('808087c8-5b4f-42f5-8998-a796ab97897d', 'Optimize database queries', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-04-09 08:10:53', 1, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('87894de4-228c-4c1c-99d3-d3564f8d58c6', 'Initialized Git repo and pushed to GitHub', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-03-31 01:09:10', 2, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('a7daf97d-6169-4725-92ad-17fadd9ecd1a', 'Draft updated wireframes for desktop and mobile views', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-04-09 08:13:54', 1, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('a832518b-c9ae-4528-8999-bb36676843ff', 'Create README file with setup instructions', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-03-31 01:07:33', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('bd30e157-fc75-4a17-a5c6-e733839b15da', 'Create user role handling logic (e.g., ADMIN, USER)', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-03-31 01:08:09', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('bdedd710-17bc-4e33-918c-e3fd39098b89', 'Built basic HTML login and registration forms', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-03-31 01:09:33', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('bff355d5-5c4c-4d38-9724-59e40b877cd1', 'Implement UserDetailsService for login', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-03-31 01:08:30', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('c39904e0-7ee8-49d2-a584-ed4889e37263', 'Style login and registration pages', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-03-31 01:08:41', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('c3b5c67a-3e16-4440-9c4a-89951353ba54', 'Integrate WebSocket for chat', '0e98413c-0078-4bdb-be9c-b31c6c7097b0', '2025-04-09 08:10:11', 1, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('c60c0c87-b31c-4d8a-b01f-54c895f9d8a6', 'Add basic validation to login/register forms', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-03-31 01:07:41', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('cca8ed73-66e2-41f5-97eb-b0c8b796d64e', 'Set up Spring Boot project with security starter', '486b539c-33d5-4518-8295-6019b1dc8816', '2025-03-31 01:09:18', 2, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('da98d938-fd4a-45f0-9bda-62c49fa8b92a', 'Test login/logout flow with dummy data', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-03-31 01:08:58', 0, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5'),
('db52c52f-752c-4cc9-9d2f-92992d2f2338', 'Set up Firebase authentication', 'de2720d6-4758-4a4b-904b-a339fc2b2187', '2025-04-09 08:10:44', 2, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('dde983f2-d4db-468a-8263-51a8dd28034d', 'Set up Firebase authentication', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-04-09 08:09:30', 2, '65041d64-f69f-48cb-9b23-8b8796debac7'),
('e067bd52-89ff-4e28-b774-57367643169e', 'Refactor sidebar navigation into collapsible drawer', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-04-09 08:14:48', 1, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('e7034030-dfd0-4172-9996-261b917dc8c2', 'Create new color scheme and typography guidelines', '0e98413c-0078-4bdb-be9c-b31c6c7097b0', '2025-04-09 08:13:29', 2, '5d4cd560-7ca7-4f59-ad2b-e6d58deb0ff8'),
('fa4ebce8-7470-46f5-9eaa-e723356a364d', 'Configured .env or application.properties', '03abaf8e-9801-43be-bf90-8f610e5a1137', '2025-03-31 01:09:43', 2, 'cacd9d9c-7e4f-45a5-aa1e-78853bd49bb5');
