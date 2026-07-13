import requests


def submit_run(base_url, profile_id, run_id, level_name, score):
    response = requests.post(
        f"{base_url}/profile/{profile_id}/run/{run_id}",
        json={
            "levelName": level_name,
            "asOf": "2026-07-13T12:00:00Z",
            "duration": 42.5,
            "score": score,
        },
        timeout=5,
    )
    response.raise_for_status()
    return response


def get_leaderboard(base_url, leaderboard_name):
    response = requests.get(f"{base_url}/leaderboard/{leaderboard_name}", timeout=5)
    response.raise_for_status()
    return response.json().get("entries") or []


def test_score_only_updates_on_higher_score(base_url, unique_id):
    leaderboard_name = f"level-{unique_id}"
    profile_id = f"user-{unique_id}"

    submit_run(base_url, profile_id, "run-1", leaderboard_name, 100)
    entries = get_leaderboard(base_url, leaderboard_name)
    assert entries == [{"profileId": profile_id, "score": 100, "rank": 1}]

    # a lower score must not overwrite the existing best score
    submit_run(base_url, profile_id, "run-2", leaderboard_name, 50)
    entries = get_leaderboard(base_url, leaderboard_name)
    assert entries == [{"profileId": profile_id, "score": 100, "rank": 1}]

    # a higher score must overwrite it
    submit_run(base_url, profile_id, "run-3", leaderboard_name, 150)
    entries = get_leaderboard(base_url, leaderboard_name)
    assert entries == [{"profileId": profile_id, "score": 150, "rank": 1}]


def test_multiple_users_are_ranked_by_score_descending(base_url, unique_id):
    leaderboard_name = f"level-{unique_id}"
    user_a = f"user-a-{unique_id}"
    user_b = f"user-b-{unique_id}"
    user_c = f"user-c-{unique_id}"

    submit_run(base_url, user_a, "run-1", leaderboard_name, 1000)
    submit_run(base_url, user_b, "run-1", leaderboard_name, 1500)
    submit_run(base_url, user_c, "run-1", leaderboard_name, 500)

    entries = get_leaderboard(base_url, leaderboard_name)

    assert entries == [
        {"profileId": user_b, "score": 1500, "rank": 1},
        {"profileId": user_a, "score": 1000, "rank": 2},
        {"profileId": user_c, "score": 500, "rank": 3},
    ]
