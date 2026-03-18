export const NotifyDonePlugin = async ({ $ }: { $: any }) => {
  return {
    event: async ({ event }: { event: { type: string } }) => {
      if (event.type === "session.idle") {
        await $`bash .opencode/scripts/notify-done.sh`;
      }
    },
  };
};
