import type { PageServerLoad } from './$types';
import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

// load function 
export const load: PageServerLoad = async ({ fetch, locals }) => {
    const response = await fetch(`${BACKEND_URL}/user/profile`, {
        headers: {
        'Authorization': `Bearer ${locals.authToken}`
        }
    });
    
    if (!response.ok) {
        throw redirect(303, `${BACKEND_URL}/auth/google`);
    }

    const data = await response.json();

    return {
        email: data.email,
        applicationsCount: data.applicationsCount,
        analytics: {
            application_velocity: data.application_velocity,
            application_velocity_trend: data.application_velocity_trend,
            resume_effectiveness: data.resume_effectiveness,
            resume_effectiveness_trend: data.resume_effectiveness_trend,
            monthly_trends: data.monthly_trends,
            interview_effectiveness: data.interview_effectiveness,
            interview_effectiveness_trend: data.interview_effectiveness_trend,
            avg_response_time: data.avg_response_time,
            avg_response_time_trend: data.avg_response_time_trend,
            rejected_count: data.rejected_count,
            ghosted_count: data.ghosted_count,
            applied_count: data.applied_count,
            screen_count: data.screen_count,
            interviewing_count: data.interviewing_count,
            offer_count: data.offer_count,
            last_updated: data.last_updated,
        }
    };
};

export const actions = {
    delete: async ({ fetch, locals }) => {
        const response = await fetch(`${BACKEND_URL}/user/deleteUser`, {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${locals.authToken}`
            },
            body: JSON.stringify({}),
        });

        if (!response.ok) {
            throw new Error('Failed to delete user');
        }
    },
} satisfies Actions;