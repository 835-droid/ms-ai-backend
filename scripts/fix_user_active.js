// MongoDB script to fix user is_active status
// Run this in mongosh or mongo shell

// Fix all users - set is_active to true
print("\n=== Fixing User Active Status ===");

// Update all users to set is_active: true
const result = db.users.updateMany(
  { is_active: { $ne: true } },
  { $set: { is_active: true, updated_at: new Date() } }
);

print(`Modified ${result.modifiedCount} user(s)`);

// Show all users with their active status
print("\n=== Current User Status ===");
db.users.find({}, {
    _id: 1,
    username: 1,
    is_active: 1,
    roles: 1
}).forEach(user => {
    print(`${user.username}: is_active=${user.is_active}, roles=${JSON.stringify(user.roles)}`);
});