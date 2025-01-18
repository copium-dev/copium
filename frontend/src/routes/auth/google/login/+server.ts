import { redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

export async function POST() {
    throw redirect(303, `${BACKEND_URL}/auth`);
}