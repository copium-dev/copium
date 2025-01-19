import { redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

export async function GET() {
    throw redirect(303, `${BACKEND_URL}/auth/google/logout`);
}