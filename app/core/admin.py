from django.contrib import admin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.utils.translation import gettext_lazy as _

from core import models

class UserAdmin(BaseUserAdmin):
    ordering = ['id']
    list_display = [
        'email', 'username', 'user_type',
        'image', 'created_at', 'updated_at',
    ]
    fieldsets = (
        (None, {'fields': ('image', 'email', 'username', 'password')}),
        (
            _('Premium Plans'),
            {
                'fields': (
                    'user_type',
                    'is_premium',
                )
            }
        ),
        (
            _('OAuth Informations'),
            {
                'fields': (
                    'is_oauth_user',
                    'oauth_type'
                )
            }
        ),
        (
            _('Tokens'),
            {
                'fields': (
                    'fcm_token',
                    'refresh_token'
                )
            }
        ),
        (
            _('User Permissions'),
            {
                'fields': (
                    'is_mail_notification_allowed',
                    'is_app_notification_allowed',
                )
            }
        ),
        (
            _('Permissions'),
            {
                'fields': (
                    'is_active',
                    'is_staff',
                    'is_superuser',
                )
            }
        ),
        (_('Date Info'), {'fields': ('created_at', 'updated_at')}),
    )
    readonly_fields = [
        'fcm_token', 'refresh_token', 'is_oauth_user',
        'oauth_type', 'created_at', 'updated_at',
    ]
    add_fieldsets = (
        (None, {
            'classes': ('wide',),
            'fields': (
                'image',
                'email',
                'username',
                'password1',
                'password2',
            ),
        }),
        (
            _('Premium Plans'),
            {
                'fields': (
                    'user_type',
                    'is_premium',
                )
            }
        ),
        (
            _('Permissions'),
            {
                'fields': (
                    'is_active',
                    'is_staff',
                    'is_superuser',
                )
            }
        ),
    )
    list_filter = ('is_premium', 'is_active', 'is_staff', 'is_superuser',)
    search_fields = ('email', 'username')


admin.site.register(models.User, UserAdmin)
