// Package builder provides helper operators used when building SQL
// filter expressions. The comment below maps common SQL operators to the
// corresponding Go method names used by this package and gives a short
// example for each.
//
// Operator        Description                          Go Method   Example
// =               Equal                                Eq         WHERE age = 30
// <> / !=         Not equal                            Neq        WHERE status <> 'inactive'
// >               Greater than                         Gt         WHERE score > 80
// <               Less than                             Lt         WHERE price < 10000
// >=              Greater than or equal                 Gte        WHERE quantity >= 10
// <=              Less than or equal                    Lte        WHERE created_at <= '2024-01-01'
// IS NULL         Is null                               IsNull     WHERE deleted_at IS NULL
// IS NOT NULL     Is not null                           IsNotNull  WHERE updated_at IS NOT NULL
// LIKE            Pattern match                         Like       WHERE name LIKE '%rama%'
// NOT LIKE        Not pattern                            NotLike    WHERE email NOT LIKE '%@spam.com'
// IN              In set                                 In         WHERE status IN ('active', 'pending')
// NOT IN          Not in set                             NotIn      WHERE id NOT IN (1,2,3)
// BETWEEN         Between                                Between    WHERE age BETWEEN 18 AND 30
// NOT BETWEEN     Not between                            NotBetween WHERE birth_year NOT BETWEEN 1990 AND 2000
// EXISTS          Exists subquery                        Exists     WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)
// NOT EXISTS      Not exists                             NotExists  WHERE NOT EXISTS (SELECT 1 FROM logs l WHERE l.user_id = u.id)
// REGEXP / ~      Regex match                             Regexp     WHERE username REGEXP '^[a-z]+'
// NOT REGEXP      Not regex                               NotRegexp  WHERE username NOT REGEXP '^[0-9]+$'
// ILIKE           Case-insensitive like                  ILike      WHERE name ILIKE '%rama%'
// NOT ILIKE       Case-insensitive not like              NotILike   WHERE email NOT ILIKE '%gmail%'
// = ANY()         Equal any (Postgres array)              EqAny      WHERE 5 = ANY(numbers)
// = ALL()         Equal all                               EqAll      WHERE score = ALL(scores)

package builder

// See each operator's implementation in this package for usage details
// and examples. The methods named above (Eq, Neq, Gt, Lt, etc.) are
// helpers to construct filter expressions in a fluent builder style.
