"""
Tests for models.
"""
from django.test import TestCase
from django.contrib.auth import get_user_model

from core import models

def create_user(email='user@example.com', password='testpass123'):
    """Create a return a new user."""
    return get_user_model().objects.create_user(email, password)

class ModelTests(TestCase):

    def test_create_user_with_email_successful(self):
        """Test creating a user with an email is successful."""
        email = 'test@example.com'
        username = 'test'
        password = 'testpass123'
        user = get_user_model().objects.create_user(
            email=email,
            username=username,
            password=password,
        )

        self.assertEqual(user.email, email)
        self.assertEqual(user.username, username)
        self.assertTrue(user.check_password(password))

    def test_new_user_email_normalized(self):
        """Test email is normalized for new users."""
        sample_emails = [
            ['test1@EXAMPLE.com', 'test1@example.com'],
            ['Test2@Example.com', 'Test2@example.com'],
            ['TEST3@EXAMPLE.com', 'TEST3@example.com'],
            ['test4@example.COM', 'test4@example.com'],
        ]
        for email, expected in sample_emails:
            user = get_user_model().objects.create_user(
                email, ('sample' + email), 'sample123'
            )
            self.assertEqual(user.email, expected)

    def test_new_user_with_extra_data(self):
        """Test creating a user with extra data"""
        email = 'test@example.com'
        username = 'test'
        password = 'testpass123'
        fcm_token = """
            erI81pGdTqC4nUZnHef9wC:APA91bEPIR_EZtwj9_q7kPlXo7xIM19ry4Kd0c_sfrEVdi6nZ-
            w8Q2THxKcuQsv1OCirvnSNVpv_gMNhhBV85R7L19oCuXxQSRLtjidQG9D4qSErjhI8ORZoso1HbfDdau8pyqUls5z3
        """

        user = get_user_model().objects.create_user(
            email=email,
            username=username,
            password=password,
            fcm_token=fcm_token
        )

        self.assertEqual(user.email, email)
        self.assertEqual(user.username, username)
        self.assertTrue(user.check_password(password))
        self.assertEqual(user.user_type, models.User.UserType.Free)
        self.assertFalse(user.is_premium)
        self.assertEqual(user.fcm_token, fcm_token)
        self.assertFalse(user.is_superuser)
        self.assertFalse(user.is_staff)
        self.assertEqual(user.refresh_token, None)

    def test_new_user_without_email_raises_error(self):
        """Test that creating a user without an email raises a ValueError."""
        with self.assertRaises(ValueError):
            get_user_model().objects.create_user('', 'test123')

    def test_create_superuser(self):
        """Test creating a superuser."""
        user = get_user_model().objects.create_superuser(
            'test@example.com',
            'test123',
            'test123',
        )

        self.assertTrue(user.is_superuser)
        self.assertTrue(user.is_staff)
