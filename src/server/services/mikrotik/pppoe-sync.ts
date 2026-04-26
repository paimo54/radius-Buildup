import { MikroTikConnection } from './client';

export async function syncMikrotikPppoeSecret(
  action: 'add' | 'set' | 'remove',
  router: { ipAddress: string; username: string; password: string; port?: number },
  user: {
    username: string;
    password?: string;
    profile?: string;
    comment?: string;
    disabled?: boolean;
  }
) {
  if (!router || !router.ipAddress || !router.username) {
    return false;
  }

  const client = new MikroTikConnection({
    host: router.ipAddress,
    username: router.username,
    password: router.password,
    port: router.port || 8728,
    timeout: 5000,
  });

  try {
    await client.connect();

    // Check if secret exists
    const existing = await client.execute('/ppp/secret/print', [`?name=${user.username}`]);
    const exists = existing && existing.length > 0;

    if (action === 'remove') {
      if (exists) {
        await client.execute('/ppp/secret/remove', [`=.id=${existing[0]['.id']}`]);
      }
    } else {
      const args = [
        `=name=${user.username}`,
        ...(user.password ? [`=password=${user.password}`] : []),
        ...(user.profile ? [`=profile=${user.profile}`] : []),
        ...(user.comment ? [`=comment=${user.comment}`] : []),
        `=disabled=${user.disabled ? 'yes' : 'no'}`,
      ];

      if (exists) {
        await client.execute('/ppp/secret/set', [`=.id=${existing[0]['.id']}`, ...args]);
      } else {
        await client.execute('/ppp/secret/add', [...args, `=service=pppoe`]);
      }
    }

    await client.disconnect();
    return true;
  } catch (error) {
    console.error(`[MikroTik PPPoE Sync] Failed for ${user.username}:`, error);
    return false;
  }
}
