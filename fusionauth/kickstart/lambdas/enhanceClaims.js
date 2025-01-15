function populate(jwt, user, registration) {
  jwt.role = 'app_user';
  if (jwt.roles && jwt.roles.includes('admin')) {
    jwt.role = 'app_admin';
  }
}
